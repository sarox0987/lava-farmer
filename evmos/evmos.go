package evmos

import (
	"context"
	"fmt"
	"lava-farmer/pkg"
	"log"
	"math/big"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	baseFee = big.NewInt(577500000000001)
	minFund = big.NewInt(0).Mul(big.NewInt(20), baseFee)
)

type network struct {
	w  *pkg.Wallet
	ps []*ethclient.Client
}

func NewNetwork(urls []string, w *pkg.Wallet) *network {
	n := &network{
		w: w,
	}
	for _, url := range urls {
		c, err := ethclient.Dial(url)
		if err != nil {
			panic(err)
		}
		n.ps = append(n.ps, c)
	}
	return n
}

func (n *network) Wallets() []string {
	ws := []string{}
	for _, a := range n.w.Accounts() {
		ws = append(ws, a.Address.Hex())
	}
	return ws
}

func (n *network) Run() {
	chainId, err := n.provider().ChainID(context.Background())
	for _, a := range n.w.Accounts() {
		if err != nil {
			fmt.Printf("Failed to retrieve chainID: %v\n", err)
			continue
		}
		go n.launch(a, chainId)
	}
}

func (n *network) Name() string {
	return "EVMOS"
}

func (n *network) launch(a accounts.Account, chainId *big.Int) {

	amount := big.NewInt(1)
	pk, _ := n.w.PrivateKey(a)

	if !n.enoughFunds(a, minFund, chainId.Int64()) {
		return
	}

	for {
		time.Sleep(20 * time.Second)
		if !n.enoughFunds(a, minFund, chainId.Int64()) {
			continue
		}
		nonce, err := n.provider().NonceAt(context.Background(), a.Address, nil)
		if err != nil {
			log.Printf("Failed to retrieve nonce: %v", err)
		}

		gasPrice, err := n.provider().SuggestGasPrice(context.Background())
		if err != nil {
			log.Printf("Failed to suggest gas price: %v", err)
			continue
		}

		gasLimit := uint64(22000)

		tx := types.NewTransaction(nonce, a.Address, amount, gasLimit, gasPrice, nil)

		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), pk)
		if err != nil {
			log.Printf("Failed to sign transaction: %v", err)
			continue
		}

		err = n.provider().SendTransaction(context.Background(), signedTx)
		if err != nil {
			log.Printf("Failed to send transaction: %d %s %v", chainId, a.Address, err)
			continue
		}

		fmt.Printf("ChainId: %d Address: %s Nonce: %d\n", chainId, a.Address, nonce)

		for {
			_, isPending, err := n.provider().TransactionByHash(context.Background(), signedTx.Hash())
			if err != nil {
				fmt.Println(err)
			}
			if !isPending {
				break
			}
		}
	}
}

func (n *network) enoughFunds(a accounts.Account, amount *big.Int, chainId int64) bool {
	balance, err := n.provider().BalanceAt(context.Background(), a.Address, nil)
	if err != nil {
		log.Printf("Failed to retrieve balance: %v", err)
		return false
	}

	if balance.Cmp(amount) == -1 {
		log.Printf("insufficient funds for: %d %s", chainId, a.Address)
		return false
	}
	return true
}

func (n *network) provider() *ethclient.Client {
	time.Sleep(60 * time.Second)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return n.ps[r.Intn(len(n.ps))]
}
