package pkg

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

type Wallet struct {
	*hdwallet.Wallet
}

func NewWallet(mn string) *Wallet {
	if mn == "" {
		mn, _ = hdwallet.NewMnemonic(128)
	}
	w, err := hdwallet.NewFromMnemonic(mn)
	if err != nil {
		log.Fatal(err)
	}

	wallet := &Wallet{w}
	for i := 0; i < 20; i++ {
		wallet.Account(i)
	}

	return wallet
}

func (w *Wallet) Account(index int) accounts.Account {
	path, _ := accounts.ParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%d", index))
	acc, _ := w.Derive(path, true)
	return acc
}
