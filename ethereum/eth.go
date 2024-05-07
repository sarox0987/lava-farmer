package ethereum

import (
	"context"
	"fmt"
	"lava-farmer/ethereum/erc20"
	"lava-farmer/pkg"
	"math/big"
	"math/rand"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
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
	w   *pkg.Wallet
	p   *ethclient.Client
	ctx context.Context
	r   *rand.Rand
}

func NewNetwork(url string, w *pkg.Wallet) *network {
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

	n := &network{
		w:   w,
		ctx: ctx,
		r:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	p, err := ethclient.Dial(url)
	if err != nil {
		panic(err)
	}
	n.p = p
	return n
}

func (n *network) Name() string { return "eth" }
func (n *network) Wallets() []string {
	return []string{}
}

func (n *network) Run() {
	for {
		for _, a := range n.w.Accounts() {
			bal, err := n.provider().BalanceAt(n.ctx, a.Address, nil)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("balance: ", a.Address, bal)
			bn, err := n.provider().BlockNumber(n.ctx)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("blockNumber: ", a.Address, bn)

			cid, err := n.provider().ChainID(n.ctx)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("chainID: ", a.Address, cid)

			nonce, err := n.provider().NonceAt(n.ctx, a.Address, nil)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("nonce: ", a.Address, nonce)
			for _, t := range tokens {
				balT, err := n.erc20BalanceOf(t, a.Address)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println("tokenBalance: ", a.Address, balT)
			}

			time.Sleep(1 * time.Minute)
		}
	}
}
func (n *network) provider() *ethclient.Client {
	t := n.r.Intn(180) + 1
	time.Sleep(time.Duration(t) * time.Second)
	return n.p
}

func (n *network) erc20BalanceOf(tokenAddress common.Address, accountAddress common.Address) (*big.Int, error) {

	c, err := erc20.NewContracts(tokenAddress, n.provider())
	if err != nil {
		return nil, err
	}

	balance, err := c.BalanceOf(&bind.CallOpts{Context: n.ctx}, accountAddress)
	if err != nil {
		return nil, err
	}

	return balance, nil
}
