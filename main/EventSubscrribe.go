package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var contractAbi = `[{"inputs":[{"internalType":"string","name":"initMessage","type":"string"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"oldStr","type":"string"},{"indexed":false,"internalType":"string","name":"newStr","type":"string"}],"name":"UpdatedMessages","type":"event"},{"inputs":[],"name":"message","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"string","name":"newMessage","type":"string"}],"name":"update","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

func main() {
	client, err := ethclient.Dial("wss://eth-rinkeby.alchemyapi.io/v2/V2pb1iVDvO15kJiisJoWLfJdi8mLWUFq")
	if err != nil {
		log.Fatal(err)
	}

	contractAddress := common.HexToAddress("0x7b053eaca2d793502157c6b20cee29f3c4fdb9ab")
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	logs, err := client.FilterLogs(context.Background(), query)
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

		err, _ := contractAbi.Unpack("Transfer", vLog.Data)
		if err != nil {
			log.Fatal(err)
		}

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
		}
	}
}
