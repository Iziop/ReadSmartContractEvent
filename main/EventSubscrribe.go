package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"strings"
	"time"
)

var contractAbi = `[{"inputs":[{"internalType":"string","name":"initMessage","type":"string"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"oldStr","type":"string"},{"indexed":false,"internalType":"string","name":"newStr","type":"string"}],"name":"UpdatedMessages","type":"event"},{"inputs":[],"name":"message","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"string","name":"newMessage","type":"string"}],"name":"update","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

//believe that amount of owner' tokens will not be too much
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

func main() {
	client, err := ethclient.Dial("wss://eth-rinkeby.alchemyapi.io/v2/V2pb1iVDvO15kJiisJoWLfJdi8mLWUFq")
	if err != nil {
		log.Fatal(err)
	}

	contractAddress := common.HexToAddress("0x7b053eaca2d793502157c6b20cee29f3c4fdb9ab")

	var lastReadBlock = int64(11257)

	var m = make(map[string][]*big.Int)

	for {

		headerByNumber, err := client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(lastReadBlock)
		fmt.Println(headerByNumber.Number)

		queryFilter := ethereum.FilterQuery{
			FromBlock: big.NewInt(lastReadBlock + 1),
			ToBlock:   headerByNumber.Number,
			Addresses: []common.Address{contractAddress},
		}

		logs, err := client.FilterLogs(context.Background(), queryFilter)
		if err != nil {
			log.Fatal(err)
		}

		contractAbi, err := abi.JSON(strings.NewReader(contractAbi))
		if err != nil {
			log.Fatal(err)
		}

		type LogTransfer struct {
			From   common.Address
			To     common.Address
			Tokens *big.Int
		}
		logTransferSig := []byte("Transfer(address,address,uint256)")
		logTransferSigHash := crypto.Keccak256Hash(logTransferSig)

		for _, vLog := range logs {
			switch vLog.Topics[0].Hex() {
			case logTransferSigHash.Hex():
				fmt.Printf("Log Name: Transfer\n")

				var transferEvent LogTransfer

				err, _ := contractAbi.Unpack("Transfer", vLog.Data)
				if err != nil {
					log.Fatal(err)
				}

				transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
				transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())
				fmt.Printf("From: %s\n", transferEvent.From.Hex())
				fmt.Printf("To: %s\n", transferEvent.To.Hex())
				fmt.Printf("Hex TokenId: %s\n", vLog.Topics[3].Big())

				if transferEvent.From.Hex() == "0x0000000000000000000000000000000000000000" {
					m[transferEvent.To.Hex()] = append(m[transferEvent.To.Hex()], vLog.Topics[3].Big())
				} else {
					m[transferEvent.From.Hex()] = findAndDelete(m[transferEvent.From.Hex()], vLog.Topics[3].Big())
					m[transferEvent.To.Hex()] = append(m[transferEvent.To.Hex()], vLog.Topics[3].Big())
				}
			}
		}
		lastReadBlock = headerByNumber.Number.Int64()
		fmt.Println("map:", m)
		time.Sleep(5 * time.Minute)
	}

}
