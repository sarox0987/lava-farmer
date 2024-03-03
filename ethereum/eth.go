package ethereum

import (
	"context"
	"lava-farmer/ethereum/erc20"
	"lava-farmer/pkg"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	tokens = []common.Address{
		common.HexToAddress("0xc3F03342514EB8362C9b6314A6974cFE698d8c98"),
		common.HexToAddress("0xF059aFA5239eD6463a00FC06a447c14Fe26406e1"),
		common.HexToAddress("0xf4DD747f81eb3A67997a117d41abb155F9cc8227"),
		common.HexToAddress("0x23687D9d40F9Ecc86E7666DDdB820e700F954526"),
		common.HexToAddress("0xf44ad89bcb12fbe8910def9f9529ce91885ad99d"),
		common.HexToAddress("0xD83dfE003E7c42077186D690DD3D24a0c965ca4e"),
		common.HexToAddress("0xe5feeaC09D36B18b3FA757E5Cf3F8dA6B8e27F4C"),
		common.HexToAddress("0x0B498ff89709d3838a063f1dFA463091F9801c2b"),
		common.HexToAddress("0xCf99D5465D39695162CA65bC59190fD92fa8e218"),
		common.HexToAddress("0xb9dfc3abb15916299eE4f51724063DcB0A1741d4"),
		common.HexToAddress("0x4F3e7F98aA70A3B879101B23b46dB1C422f85F52"),
		common.HexToAddress("0x68ab9f9ba8E490D9Dc191f7A90540b5EDCcC92D6"),
		common.HexToAddress("0xD13bF4acF5d4b2407d785d2528D746fA75CF9778"),
		common.HexToAddress("0xEab53D2Dc181d5c2D316a3370471a028C621EdE7"),
		common.HexToAddress("0xfe05b972eab7761b60b4a14558eac3fef78f306a"),
		common.HexToAddress("0xe4e6570c098B4C0E780E94fa10e1B3b23F5E28a8"),
		common.HexToAddress("0x693391144D1e079e20cC64f795e9450C94966171"),
	}
)

type network struct {
	w *pkg.Wallet
	c *ethclient.Client
}

func NewNetwork(url string, w *pkg.Wallet) *network {
	n := &network{
		w: w,
	}

	c, err := ethclient.Dial(url)
	if err != nil {
		panic(err)
	}
	n.c = c
	return n
}

func (n *network) Name() string { return "eth" }
func (n *network) Wallets() []string {
	return []string{}
}

func (n *network) Run() {
	for {
		for _, a := range n.w.Accounts() {
			n.c.BalanceAt(context.Background(), a.Address, nil)
			n.c.BlockNumber(context.Background())
			n.c.ChainID(context.Background())
			n.c.NonceAt(context.Background(), a.Address, nil)
			for _, t := range tokens {
				n.erc20BalanceOf(t, a.Address)
			}
			time.Sleep(10 * time.Second)
		}
	}
}

func (n *network) erc20BalanceOf(tokenAddress common.Address, accountAddress common.Address) (*big.Int, error) {

	c, err := erc20.NewContracts(tokenAddress, n.c)
	if err != nil {
		return nil, err
	}

	balance, err := c.BalanceOf(nil, accountAddress)
	if err != nil {
		return nil, err
	}

	return balance, nil
}
