package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"lava-farmer/ethereum"
	"lava-farmer/pkg"
	"lava-farmer/stark"
	"os"
)

type env struct {
	EthTestnet   string `json:"eth_testnet"`
	EvmosMainnet string `json:"evmos_mainnet"`
	EvmosTestnet string `json:"evmos_testnet"`
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

	// evmosT := evmos.NewNetwork(e.EvmosTestnet, w)
	starkT := stark.NewNetwork(e.StarkTestnet)
	ns = append(ns, starkT)

	for _, n := range ns {
		fmt.Println("Fund These Addresses: ")
		fmt.Println()
		fmt.Printf("%s: \n", n.Name())
		fmt.Printf("  {\n")

		for _, a := range n.Wallets() {
			fmt.Printf("\t%s\n", a)
		}
		fmt.Printf("  }\n\n")
	}

	ethT := ethereum.NewNetwork(e.EthTestnet, w)

	fmt.Print("Enter >>> ")

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
	go starkT.Run()
	go ethT.Run()
	select {}
}
