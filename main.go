package main

import (
	"encoding/json"
	"io"
	"lava-farmer/ethereum"
	"lava-farmer/pkg"
	"lava-farmer/stark"
	"os"
)

type env struct {
	EthMainnet   string `json:"eth_mainnet"`
	StarkMainnet string `json:"stark_mainnet"`
	StarkTestnet string `json:"stark_testnet"`
	Mnemonic     string `json:"mnemonic"`
}

type network interface {
	Wallets() []string
	Run()
	Name() string
}

func main() {

	ns := []network{}
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

	w := pkg.NewWallet(e.Mnemonic)
	var starkT, starkM, ethM network

	if len(e.StarkTestnet) > 0 {
		starkT = stark.NewNetwork(e.StarkTestnet)
		ns = append(ns, starkT)
	}
	if len(e.StarkMainnet) > 0 {
		starkM = stark.NewNetwork(e.StarkMainnet)
		ns = append(ns, starkM)
	}

	if len(e.EthMainnet) > 0 {
		ethM = ethereum.NewNetwork(e.EthMainnet, w)
		ns = append(ns, ethM)
	}

	for _, n := range ns {
		if n != nil {
			go n.Run()
		}
	}

	select {}
}
