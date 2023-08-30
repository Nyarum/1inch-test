package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"strings"

	store "1inch/store"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

const (
	UniswapFactoryAddr = "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"
)

func weiToDecimal(wei *big.Int, delim *big.Float) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), delim)
}

func etherToWei(eth *big.Float) *big.Int {
	truncInt, _ := eth.Int(nil)
	truncInt = new(big.Int).Mul(truncInt, big.NewInt(params.Ether))
	fracStr := strings.Split(fmt.Sprintf("%.18f", eth), ".")[1]
	fracStr += strings.Repeat("0", 18-len(fracStr))
	fracInt, _ := new(big.Int).SetString(fracStr, 10)
	wei := new(big.Int).Add(truncInt, fracInt)
	return wei
}

// getAmountOut reimplements math from uniswap contract
func getAmountOut(inputAmount int, reserve0, reserve1 *big.Int) *big.Float {
	amountInWithFee := inputAmount * 997
	numerator := new(big.Int).Mul(big.NewInt(int64(amountInWithFee)), reserve1)
	preDenominator := new(big.Int).Mul(reserve0, big.NewInt(1000))
	denominator := new(big.Int).Add(preDenominator, big.NewInt(int64(amountInWithFee)))

	return new(big.Float).Quo(new(big.Float).SetInt(numerator), new(big.Float).SetInt(denominator))
}

func main() {
	uniswapFactoryAddr := flag.String("factory-addr", UniswapFactoryAddr, "Put address of uniswap factory")
	inputToken := flag.String("input-token", "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", "Put address of input token (default: WETH)")
	outputToken := flag.String("output-token", "0xdac17f958d2ee523a2206206994597c13d831ec7", "Put address of output token (default: USDT)")
	inputAmount := flag.Int("input-amount", 1, "Input amount to estimate output price")
	flag.Parse()

	client, err := ethclient.Dial("https://mainnet.infura.io/v3/06eaf9e210cd4587a85c1666dd1b2c17")
	if err != nil {
		log.Fatal(err)
	}

	uniswapV2Factory, err := store.NewUniswapv2factory(common.HexToAddress(*uniswapFactoryAddr), client)
	if err != nil {
		log.Fatal(err)
	}

	pairAddr, err := uniswapV2Factory.GetPair(
		&bind.CallOpts{},
		common.HexToAddress(*inputToken),  // WETH
		common.HexToAddress(*outputToken), // USDT
	)
	if err != nil {
		log.Fatal(err)
	}

	uniswapV2Pair, err := store.NewUniswapv2pair(pairAddr, client)
	if err != nil {
		log.Fatal(err)
	}

	reserves, err := uniswapV2Pair.GetReserves(&bind.CallOpts{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(etherToWei(getAmountOut(*inputAmount, reserves.Reserve0, reserves.Reserve1)))
}
