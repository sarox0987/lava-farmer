package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

var (
	baseFee = big.NewInt(577500000000001)
	minFund = big.NewInt(0).Mul(big.NewInt(20), baseFee)
)

type env struct {
	MainnetUrl string `json:"mainnet_url"`
	TestnetUrl string `json:"testnet_url"`
	Mnemonic   string `json:"mnemonic"`
}

func main() {

	file, err := os.Open("env.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	e := &env{}
	if err := json.Unmarshal(data, e); err != nil {
		panic(err)
	}

	w := newWallet(e.Mnemonic)
	for _, a := range w.Accounts() {
		fmt.Println(a.URL.Path, a.Address)
	}
	fmt.Print("Enter >>> ")

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	urls := []string{e.MainnetUrl, e.TestnetUrl}

	for _, a := range w.Accounts() {
		for _, u := range urls {
			client, err := ethclient.Dial(u)
			if err != nil {
				log.Fatalf("Failed to connect to the client: %v", err)
			}
			chainId, err := client.ChainID(context.Background())
			if err != nil {
				log.Fatalf("Failed to retrieve chainID: %v", err)
			}
			go w.launch(client, a, chainId)
		}
	}

	select {}
}

func (w *wallet) launch(client *ethclient.Client, a accounts.Account, chainId *big.Int) {

	amount := big.NewInt(1)
	pk, _ := w.PrivateKey(a)

	if !enoughFunds(client, a, minFund, chainId.Int64()) {
		return
	}

	for {
		time.Sleep(20 * time.Second)
		if !enoughFunds(client, a, minFund, chainId.Int64()) {
			continue
		}
		nonce, err := client.NonceAt(context.Background(), a.Address, nil)
		if err != nil {
			log.Printf("Failed to retrieve nonce: %v", err)
		}

		gasPrice, err := client.SuggestGasPrice(context.Background())
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

		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			log.Printf("Failed to send transaction: %d %s %v", chainId, a.Address, err)
			continue
		}

		fmt.Printf("ChainId: %d Address: %s Nonce: %d\n", chainId, a.Address, nonce)

		for {
			time.Sleep(2 * time.Second)
			_, isPending, err := client.TransactionByHash(context.Background(), signedTx.Hash())
			if err != nil {
				time.Sleep(20 * time.Second)
			}
			if !isPending {
				break
			}
		}
	}
}

func enoughFunds(client *ethclient.Client, a accounts.Account, amount *big.Int, chainId int64) bool {
	balance, err := client.BalanceAt(context.Background(), a.Address, nil)
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

type wallet struct {
	*hdwallet.Wallet
}

func newWallet(mn string) *wallet {
	if mn == "" {
		log.Fatal("add mnemonic phrase")
	}
	w, err := hdwallet.NewFromMnemonic(mn)
	if err != nil {
		log.Fatal(err)
	}

	wallet := &wallet{w}
	for i := 0; i < 10; i++ {
		wallet.Account(i)
	}
	return wallet
}

func (w *wallet) Account(index int) accounts.Account {
	path, _ := accounts.ParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%d", index))
	acc, _ := w.Derive(path, true)
	return acc
}
