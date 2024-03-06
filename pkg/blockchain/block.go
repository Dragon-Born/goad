package blockchain

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	log "github.com/sirupsen/logrus"
	"math/big"
)

func (c *Client) GetCurrentBlock() *big.Int {
	number, err := c.client.BlockByNumber(context.Background(), nil)
	if err != nil {
		return big.NewInt(0)
	}
	return number.Number()
}

func (c *Client) GetBlockTransactions(blockNumber int64) (dexTransactions []DexTransaction, err error) {
	multiCallHex := "0x5ae401dc"
	block, err := c.client.BlockByNumber(context.Background(), big.NewInt(blockNumber))
	if err != nil {
		return
	}
	supportedID := GetSupportedList()
	for _, tx := range block.Transactions() {
		if len(tx.Data()) < 4 {
			continue
		}
		data := tx.Data()
		functionID := data[:4]
		if hexutil.Encode(functionID) == multiCallHex {
			decoded, err := DecodeTransaction(hexutil.Encode(tx.Data()))
			if err != nil {
				log.Debugf(fmt.Sprintf("Multicall Error decoding tranaction: %v, TX: %s", err, tx.Hash().Hex()))
				continue
			}
			d := decoded["data"].([][]uint8)[0]
			data = d
			functionID = data[:4]
		}
		_, ok := supportedID[hexutil.Encode(functionID)]
		if !ok {
			continue
		}
		dex, err := DecodeTransaction(hexutil.Encode(data))
		if err != nil {
			log.Debugf(fmt.Sprintf("Error decoding tranaction: %v, TX: %s", err, tx.Hash().Hex()))
			continue
		}
		value := tx.Value()
		if value.Int64() == 0 {
			value = dex["amountIn"].(*big.Int)
		}
		dexTransactions = append(dexTransactions, DexTransaction{
			From:     dex["to"].(common.Address),
			Method:   hexutil.Encode(functionID),
			Amount:   value,
			GasPrice: tx.GasPrice(),
		})
	}
	return
}
