package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const UniswapV3FactoryABI = `[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token0","type":"address"},{"indexed":true,"internalType":"address","name":"token1","type":"address"},{"indexed":true,"internalType":"uint24","name":"fee","type":"uint24"},{"indexed":false,"internalType":"int24","name":"tickSpacing","type":"int24"},{"indexed":false,"internalType":"address","name":"pool","type":"address"}],"name":"PoolCreated","type":"event"}]`

const UniswapV3FactoryAddress = "0x33128a8fC17869897dcE68Ed026d694621f6FDfD"

func main() {
	client, err := ethclient.Dial("https://mainnet.base.org")
	if err != nil {
		log.Fatalf("Failed to connect to the Base client: %v", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(UniswapV3FactoryABI))
	if err != nil {
		log.Fatalf("Failed to parse ABI: %v", err)
	}

	filterQuery := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(UniswapV3FactoryAddress)},
		Topics: [][]common.Hash{
			{parsedABI.Events["PoolCreated"].ID},
		},
	}

	var lastBlock uint64 = 0
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get the latest block number
		latestBlock, err := client.BlockNumber(ctx)
		if err != nil {
			log.Printf("Failed to get latest block number: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if lastBlock == 0 {
			lastBlock = latestBlock - 1
		}

		filterQuery.FromBlock = new(big.Int).SetUint64(lastBlock + 1)
		filterQuery.ToBlock = new(big.Int).SetUint64(latestBlock)

		logs, err := client.FilterLogs(ctx, filterQuery)
		if err != nil {
			log.Printf("Failed to filter logs: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, vLog := range logs {
			event := struct {
				Token0      common.Address
				Token1      common.Address
				Fee         *big.Int
				TickSpacing *big.Int
				Pool        common.Address
			}{}
			err := parsedABI.UnpackIntoInterface(&event, "PoolCreated", vLog.Data)
			if err != nil {
				log.Printf("Failed to unpack log: %v", err)
				continue
			}

			event.Token0 = common.HexToAddress(vLog.Topics[1].Hex())
			event.Token1 = common.HexToAddress(vLog.Topics[2].Hex())

			fmt.Printf("New pool created:\n")
			fmt.Printf("Token0: %s\n", event.Token0.Hex())
			fmt.Printf("Token1: %s\n", event.Token1.Hex())
			fmt.Printf("Fee: %s\n", event.Fee.String())
			fmt.Printf("TickSpacing: %s\n", event.TickSpacing.String())
			fmt.Printf("Pool Address: %s\n\n", event.Pool.Hex())
		}

		lastBlock = latestBlock
		time.Sleep(15 * time.Second)
	}
}
