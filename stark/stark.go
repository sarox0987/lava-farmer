package stark

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/account"
	srpc "github.com/NethermindEth/starknet.go/rpc"
	"github.com/NethermindEth/starknet.go/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	classHash, _        = utils.HexToFelt("0x2794ce20e5f2ff0d40e632cb53845b9f4e526ebd8471983f7dbd355b721d5a")
	ethContract  string = "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"
)

type starkAccount struct {
	acc      *account.Account
	ks       *account.MemKeystore
	pub      *felt.Felt
	pk       *felt.Felt
	deployed bool
}

type network struct {
	accounts map[string]starkAccount
	p        *srpc.Provider
	chainId  string
}

type faccount struct {
	PubKey   string `json:"pub_key"`
	PrvKey   string `json:"prv_key"`
	Addreess string `json:"address"`
}

func NewNetwork(url string) *network {
	file, err := os.OpenFile("stark-accounts.json", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	accs := []faccount{}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &accs); err != nil {
			panic(err)
		}
	}
	c, err := rpc.Dial(url)
	if err != nil {
		panic(err)
	}
	p := srpc.NewProvider(c)
	chianId, err := p.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	n := &network{
		accounts: make(map[string]starkAccount),
		p:        p,
		chainId:  chianId,
	}

	if len(accs) == 0 {
		for i := 0; i < 10; i++ {
			ks, pub, pk := account.GetRandomKeys()
			accountAddressFelt, _ := new(felt.Felt).SetString("0x1")
			a, err := account.NewAccount(p, accountAddressFelt, pub.String(), ks, 2)
			if err != nil {
				continue
			}

			precomputedAddress, err := a.PrecomputeAddress(&felt.Zero, pub, classHash, []*felt.Felt{pub})
			if err != nil {
				continue
			}

			a.AccountAddress = precomputedAddress
			n.accounts[precomputedAddress.String()] = starkAccount{
				acc: a,
				ks:  ks,
				pub: pub,
				pk:  pk,
			}
			accs = append(accs, faccount{
				PubKey:   pub.String(),
				PrvKey:   pk.String(),
				Addreess: precomputedAddress.String(),
			})
		}
		b, _ := json.Marshal(accs)
		_, err := file.Write(b)
		if err != nil {
			panic(err)
		}
	} else {
		for _, acc := range accs {
			addressFelt, _ := utils.HexToFelt(acc.Addreess)
			pubFelt, _ := utils.HexToFelt(acc.PubKey)
			prvFelt, _ := utils.HexToFelt(acc.PrvKey)
			ks := account.NewMemKeystore()
			fakePrivKeyBI, ok := new(big.Int).SetString(acc.PrvKey, 0)
			if !ok {
				continue
			}
			ks.Put(acc.PubKey, fakePrivKeyBI)
			a, err := account.NewAccount(p, addressFelt, acc.PubKey, ks, 2)
			if err != nil {
				continue
			}
			n.accounts[acc.Addreess] = starkAccount{
				acc:      a,
				ks:       ks,
				pub:      pubFelt,
				pk:       prvFelt,
				deployed: true,
			}
		}
	}
	return n
}

func (n *network) Name() string {
	return "STARK"
}
func (n *network) Run() {
	go func() {
		for {
			for a := range n.accounts {
				n.getBalance(a)
				time.Sleep(5 * time.Second)
			}
		}
	}()
	for k, a := range n.accounts {
		if !a.deployed {
			time.Sleep(10 * time.Second)
			go func(k string, a starkAccount) {
				for {
					time.Sleep(10 * time.Second)
					tx := srpc.DeployAccountTxn{
						Nonce:               &felt.Zero,
						MaxFee:              new(felt.Felt).SetUint64(4724395326064),
						Type:                srpc.TransactionType_DeployAccount,
						Version:             srpc.TransactionV1,
						Signature:           []*felt.Felt{},
						ClassHash:           classHash,
						ContractAddressSalt: a.pub,
						ConstructorCalldata: []*felt.Felt{a.pub},
					}
					pa, _ := utils.HexToFelt(k)
					err := a.acc.SignDeployAccountTransaction(context.Background(), &tx, pa)
					if err != nil {
						continue
					}
					resp, err := a.acc.AddDeployAccountTransaction(context.Background(), srpc.BroadcastDeployAccountTxn{DeployAccountTxn: tx})
					if err != nil {
						continue
					}
					fmt.Printf("contract %s deployed with tx: %s\n", k, resp.TransactionHash.String())
					break
				}
				go launch(a, n.chainId)

			}(k, a)
		} else {
			go launch(a, n.chainId)
		}
	}
}

func launch(a starkAccount, chainId string) {

	maxfee := new(felt.Felt).SetUint64(4783000019481)
	for {
		time.Sleep(10 * time.Second)
		nonce, err := a.acc.Nonce(context.Background(), srpc.BlockID{Tag: "latest"}, a.acc.AccountAddress)
		if err != nil {
			continue
		}
		InvokeTx := srpc.InvokeTxnV1{
			MaxFee:        maxfee,
			Version:       srpc.TransactionV1,
			Nonce:         nonce,
			Type:          srpc.TransactionType_Invoke,
			SenderAddress: a.acc.AccountAddress,
		}

		contractAddress, err := utils.HexToFelt(ethContract)
		if err != nil {
			continue
		}

		FnCall := srpc.FunctionCall{
			ContractAddress:    contractAddress,
			EntryPointSelector: utils.GetSelectorFromNameFelt("transfer"),
			Calldata:           []*felt.Felt{a.acc.AccountAddress, utils.BigIntToFelt(common.Big1), utils.BigIntToFelt(common.Big0)},
		}

		InvokeTx.Calldata, err = a.acc.FmtCalldata([]srpc.FunctionCall{FnCall})
		if err != nil {
			continue
		}

		err = a.acc.SignInvokeTransaction(context.Background(), &InvokeTx)
		if err != nil {
			continue
		}

		a.acc.AddInvokeTransaction(context.Background(), InvokeTx)

	}
}

func (n *network) getBalance(accountAddress string) {
	tokenAddressInFelt, _ := utils.HexToFelt(ethContract)
	accountAddressInFelt, _ := utils.HexToFelt(accountAddress)

	tx := srpc.FunctionCall{
		ContractAddress:    tokenAddressInFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("balanceOf"),
		Calldata:           []*felt.Felt{accountAddressInFelt},
	}
	n.p.Call(context.Background(), tx, srpc.BlockID{Tag: "latest"})

	getDecimalsTx := srpc.FunctionCall{
		ContractAddress:    tokenAddressInFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("decimals"),
	}
	n.p.Call(context.Background(), getDecimalsTx, srpc.BlockID{Tag: "latest"})
}

func (n *network) Wallets() []string {
	ws := []string{}
	for k := range n.accounts {
		ws = append(ws, k)
	}
	return ws
}
