package contract

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type LogTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
}

type EventsFetcher struct {
	client          *ethclient.Client
	contractAddress common.Address
}

func NewEventsFetcher(rpcUrl string, contractAddressHex string) (*EventsFetcher, error) {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, err
	}

	contractAddress := common.HexToAddress(contractAddressHex)
	return &EventsFetcher{client: client, contractAddress: contractAddress}, nil
}

func (fetcher *EventsFetcher) FetchFrom(blockNumberFrom *big.Int) (blockNumberTo *big.Int, transferEvents []LogTransfer, err error) {
	client := fetcher.client
	contractAddress := fetcher.contractAddress

	blockNumber, err := client.BlockNumber(context.Background())
	if err != nil {
		return nil, nil, err
	}
	blockNumberTo = new(big.Int).SetUint64(blockNumber)

	if blockNumberFrom.Cmp(blockNumberTo) > 0 {
		return blockNumberTo, []LogTransfer{}, nil
	}

	logTransferSigHash := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

	queryFilter := ethereum.FilterQuery{
		FromBlock: blockNumberFrom,
		ToBlock:   blockNumberTo,
		Addresses: []common.Address{contractAddress},
		Topics:    [][]common.Hash{{logTransferSigHash}},
	}

	events, err := client.FilterLogs(context.Background(), queryFilter)
	if err != nil {
		return nil, nil, err
	}

	transferEvents = []LogTransfer{}
	for _, event := range events {
		transferEvent := LogTransfer{
			From:    common.HexToAddress(event.Topics[1].Hex()),
			To:      common.HexToAddress(event.Topics[2].Hex()),
			TokenId: event.Topics[3].Big(),
		}
		transferEvents = append(transferEvents, transferEvent)
	}

	return blockNumberTo, transferEvents, nil
}
