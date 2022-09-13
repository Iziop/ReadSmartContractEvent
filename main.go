package main

import (
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/Iziop/ReadSmartContractEvent/contract"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	rpcUrl := "wss://eth-rinkeby.alchemyapi.io/v2/V2pb1iVDvO15kJiisJoWLfJdi8mLWUFq"
	contractAddress := "0x7b053eaca2d793502157c6b20cee29f3c4fdb9ab"
	contractCreationBlockNumber := int64(11218682)

	eventFetcher, err := contract.NewEventsFetcher(rpcUrl, contractAddress)
	if err != nil {
		log.Panicf("Error: %v", err)
	}

	usersTokens := make(map[common.Address][]*big.Int)
	blockNumberFrom := big.NewInt(contractCreationBlockNumber)
	for {
		blockNumberTo, transfers, err := eventFetcher.FetchFrom(blockNumberFrom)
		if err != nil {
			log.Panicf("Error: %v", err)
		}
		fmt.Printf("Reading events from block %v to %v\n", blockNumberFrom, blockNumberTo)

		for _, transfer := range transfers {
			if _, ok := usersTokens[transfer.From]; ok {
				usersTokens[transfer.From] = findAndDelete(usersTokens[transfer.From], transfer.TokenId)
				if len(usersTokens[transfer.From]) == 0 {
					delete(usersTokens, transfer.From)
				}
			}
			usersTokens[transfer.To] = append(usersTokens[transfer.To], transfer.TokenId)
		}
		blockNumberFrom = blockNumberTo.Add(blockNumberTo, big.NewInt(1))
		printUsersTokens(usersTokens)

		time.Sleep(30 * time.Second)
	}
}

// believe that amount of owner's tokens will not be too much
func findAndDelete(s []*big.Int, item *big.Int) []*big.Int {
	index := 0
	for _, i := range s {
		if i.Cmp(item) != 0 {
			s[index] = i
			index++
		}
	}
	return s[:index]
}

func printUsersTokens(usersTokens map[common.Address][]*big.Int) {
	mapTokenIDs := func(tokenIDs []*big.Int) []string {
		result := []string{}
		for _, tokenID := range tokenIDs {
			result = append(result, tokenID.String())
		}
		return result
	}

	for address, tokens := range usersTokens {
		fmt.Printf("[%v] => [%v]\n", address, strings.Join(mapTokenIDs(tokens), ","))
	}
}
