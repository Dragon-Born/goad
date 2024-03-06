package blockchain

import (
	"errors"
	"github.com/ethereum/go-ethereum/ethclient"
)

const clientURL = "https://bsc-dataseed.binance.org/" // Example client URL, change it to your Ethereum node URL

type Client struct {
	client *ethclient.Client
}

func NewClient(rpc string) (*Client, error) {
	c, err := ethclient.Dial(rpc)
	if err != nil {
		return &Client{}, errors.New("error on connect to client: " + err.Error())
	}
	return &Client{c}, nil
}
