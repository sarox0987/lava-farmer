package stark

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net/http"
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
	Accounts map[string]starkAccount
	p        *srpc.Provider
	chainId  string
	ctx      context.Context
	r        *rand.Rand
}

type faccount struct {
	PubKey   string `json:"pub_key"`
	PrvKey   string `json:"prv_key"`
	Addreess string `json:"address"`
}

func (n *network) provider() *srpc.Provider {
	t := n.r.Intn(180) + 1
	time.Sleep(time.Duration(t) * time.Second)
	return n.p
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

	hdr := http.Header{}
	hdr.Set("accept", "application/json")
	hdr.Set("accept-language", "en-US,en;q=0.9")
	hdr.Set("content-type", "application/json")
	hdr.Set("priority", "u=1, i")
	hdr.Set("sec-ch-ua", `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`)
	hdr.Set("sec-ch-ua-mobile", "?0")
	hdr.Set("sec-ch-ua-platform", `"windows"`)
	hdr.Set("sec-fetch-dest", "empty")
	hdr.Set("sec-fetch-mode", "cors")
	hdr.Set("sec-fetch-site", "none")
	hdr.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.246")

	ctx := rpc.NewContextWithHeaders(context.Background(), hdr)

	accs := []faccount{}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &accs); err != nil {
			panic(err)
		}
	}
	var chainId string

	c, err := rpc.Dial(url)
	if err != nil {
		panic(err)
	}
	p := srpc.NewProvider(c)
	cId, err := p.ChainID(ctx)
	if err != nil {
		panic(err)
	}
	chainId = cId

	n := &network{
		Accounts: make(map[string]starkAccount),
		p:        p,
		chainId:  chainId,
		ctx:      ctx,
		r:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	// if chainId == "SN_MAIN" {
	// 	return n
	// }

	// if len(accs) == 0 {
	// 	for i := 0; i < 10; i++ {
	// 		ks, pub, pk := account.GetRandomKeys()
	// 		accountAddressFelt, _ := new(felt.Felt).SetString("0x1")
	// 		a, err := account.NewAccount(n.ps[i%len(ps)], accountAddressFelt, pub.String(), ks, 2)
	// 		if err != nil {
	// 			continue
	// 		}

	// 		precomputedAddress, err := a.PrecomputeAddress(&felt.Zero, pub, classHash, []*felt.Felt{pub})
	// 		if err != nil {
	// 			continue
	// 		}

	// 		a.AccountAddress = precomputedAddress
	// 		n.Accounts[precomputedAddress.String()] = starkAccount{
	// 			acc: a,
	// 			ks:  ks,
	// 			pub: pub,
	// 			pk:  pk,
	// 		}
	// 		accs = append(accs, faccount{
	// 			PubKey:   pub.String(),
	// 			PrvKey:   pk.String(),
	// 			Addreess: precomputedAddress.String(),
	// 		})
	// 	}
	// 	b, _ := json.Marshal(accs)
	// 	_, err := file.Write(b)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// } else {
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
		a, err := account.NewAccount(n.p, addressFelt, acc.PubKey, ks, 2)
		if err != nil {
			fmt.Println(err)
			continue
		}
		n.Accounts[acc.Addreess] = starkAccount{
			acc:      a,
			ks:       ks,
			pub:      pubFelt,
			pk:       prvFelt,
			deployed: true,
		}
	}
	// }
	return n
}

func (n *network) Name() string {
	return "STARK"
}

func (n *network) Run() {
	for {
		for a := range n.Accounts {
			n.getBalance(a)
			cid, err := n.provider().ChainID(n.ctx)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("ChainID: ", cid)
			bn, err := n.provider().BlockNumber(n.ctx)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(cid, " BlockNumber: ", bn)

			bhn, err := n.provider().BlockHashAndNumber(n.ctx)
			if err != nil {
				fmt.Println(err)
				continue
			}
			countn, err := n.provider().BlockTransactionCount(n.ctx, srpc.BlockID{
				Number: &bhn.BlockNumber,
			})
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(cid, " BlockTransactionCountWithNumber: ", countn)

			counth, err := n.provider().BlockTransactionCount(n.ctx, srpc.BlockID{
				Hash: bhn.BlockHash,
			})
			fmt.Println(cid, " BlockTransactionCountWithHash: ", counth)

			fmt.Println()
		}
	}
}

func launch(a starkAccount, chainId string, ctx context.Context) {
	maxfee := new(felt.Felt).SetUint64(4783000019481)
	for {
		time.Sleep(2 * time.Minute)
		nonce, err := a.acc.Nonce(ctx, srpc.BlockID{Tag: "latest"}, a.acc.AccountAddress)
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

		err = a.acc.SignInvokeTransaction(ctx, &InvokeTx)
		if err != nil {
			continue
		}

		res, err := a.acc.AddInvokeTransaction(ctx, InvokeTx)
		if err != nil {
			continue
		}
		fmt.Println(res.TransactionHash)
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
	f, err := n.provider().Call(n.ctx, tx, srpc.BlockID{Tag: "latest"})
	fmt.Println(n.chainId, accountAddress, f, err)

	getDecimalsTx := srpc.FunctionCall{
		ContractAddress:    tokenAddressInFelt,
		EntryPointSelector: utils.GetSelectorFromNameFelt("decimals"),
	}
	f, err = n.provider().Call(n.ctx, getDecimalsTx, srpc.BlockID{Tag: "latest"})
	fmt.Println(n.chainId, accountAddress, f, err)
}

func (n *network) Wallets() []string {
	ws := []string{}
	for k := range n.Accounts {
		ws = append(ws, k)
	}
	return ws
}
