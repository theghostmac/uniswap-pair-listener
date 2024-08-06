package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

//const UniswapV3FactoryABI = `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint24","name":"fee","type":"uint24"},{"indexed":true,"internalType":"int24","name":"tickSpacing","type":"int24"}],"name":"FeeAmountEnabled","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"oldOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnerChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token0","type":"address"},{"indexed":true,"internalType":"address","name":"token1","type":"address"},{"indexed":true,"internalType":"uint24","name":"fee","type":"uint24"},{"indexed":false,"internalType":"int24","name":"tickSpacing","type":"int24"},{"indexed":false,"internalType":"address","name":"pool","type":"address"}],"name":"PoolCreated","type":"event"},{"inputs":[{"internalType":"address","name":"tokenA","type":"address"},{"internalType":"address","name":"tokenB","type":"address"},{"internalType":"uint24","name":"fee","type":"uint24"}],"name":"createPool","outputs":[{"internalType":"address","name":"pool","type":"address"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint24","name":"fee","type":"uint24"},{"internalType":"int24","name":"tickSpacing","type":"int24"}],"name":"enableFeeAmount","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint24","name":"","type":"uint24"}],"name":"feeAmountTickSpacing","outputs":[{"internalType":"int24","name":"","type":"int24"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"address","name":"","type":"address"},{"internalType":"uint24","name":"","type":"uint24"}],"name":"getPool","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"parameters","outputs":[{"internalType":"address","name":"factory","type":"address"},{"internalType":"address","name":"token0","type":"address"},{"internalType":"address","name":"token1","type":"address"},{"internalType":"uint24","name":"fee","type":"uint24"},{"internalType":"int24","name":"tickSpacing","type":"int24"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_owner","type":"address"}],"name":"setOwner","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
//
//const UniswapV3FactoryAddress = "0x1F98431c8aD98523631AE4a59f267346ea31F984"

func main2() {
	client, err := ethclient.Dial("https://mainnet.base.org")
	if err != nil {
		log.Fatalf("Failed to connect to the Base client: %v", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(UniswapV3FactoryABI))
	if err != nil {
		log.Fatalf("Failed to parse ABI: %v", err)
	}

	// Create a filter query for the PoolCreated event.
	filterQuery := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(UniswapV3FactoryAddress)},
		Topics: [][]common.Hash{
			{
				parsedABI.Events["PoolCreated"].ID,
			},
		},
	}

	// Create a channel to receive the logs.
	logs := make(chan types.Log)

	// Now, subscribe to the logs.
	sub, err := client.SubscribeFilterLogs(context.Background(), filterQuery, logs)
	if err != nil {
		log.Fatalf("Failed to subscribe to the logs after filtering: %v", err)
	}

	fmt.Println("Listening for PoolCreated events...")

	for {
		select {
		case err := <-sub.Err():
			log.Fatalf("Failed to subscribe to the logs while listening to pool creation events: %v", err)
		case vLog := <-logs:
			// Parse these events' data.
			event := struct {
				Token0      common.Address
				Token1      common.Address
				Fee         *big.Int
				TickSpacing *big.Int
				Pool        common.Address
			}{}
			err := parsedABI.UnpackIntoInterface(&event, "PoolCreated", vLog.Data)
			if err != nil {
				log.Printf("Failed to parse PoolCreated event by unpacking into log: %v", err)
				continue
			}

			// Now we can extract the token addresses from the topics.
			event.Token0 = common.HexToAddress(vLog.Topics[1].Hex())
			event.Token1 = common.HexToAddress(vLog.Topics[2].Hex())

			// Print everything out.
			fmt.Printf("New pool created:\n")
			fmt.Printf("Token0: %s\n", event.Token0.Hex())
			fmt.Printf("Token1: %s\n", event.Token1.Hex())
			fmt.Printf("Fee: %s\n", event.Fee.String())
			fmt.Printf("TickSpacing: %s\n", event.TickSpacing.String())
			fmt.Printf("Pool Address: %s\n\n", event.Pool.Hex())
		}
	}
}
