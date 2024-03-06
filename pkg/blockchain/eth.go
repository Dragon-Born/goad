package blockchain

import (
	"context"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"goad/pkg/cache"
	"goad/pkg/wallet"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (c *Client) GetBalance(accountAddress common.Address) (*big.Int, error) {
	balance, err := c.client.BalanceAt(context.Background(), accountAddress, nil)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

type PreTransaction struct {
	nonce    uint64
	gasPrice *big.Int
	gasLimit uint64
}

func (c *Client) GetGasLimit() (uint64, error) {
	cIdentifier := fmt.Sprintf("gas_price")
	ca := cache.GetCache()
	gas, ok := ca.Get(cIdentifier)
	if ok {
		return gas.(uint64), nil
	}
	header, err := c.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	ca.Set(cIdentifier, header.GasLimit, time.Minute*60)
	spew.Dump(header)
	return header.GasLimit, nil
}

func (c *Client) CalculateCost(tokenAddress common.Address) (*big.Int, error) {
	// Suggest gas price
	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	gasLimit, err := c.client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: tokenAddress,
	})
	if err != nil {
		return nil, err
	}

	gasLimitBigInt := big.NewInt(0).SetUint64(gasLimit + (gasLimit * 20 / 100))
	totalCost := big.NewInt(0).Mul(gasPrice, gasLimitBigInt)

	return totalCost, nil
}

func (c *Client) GetPreTransaction(accountAddress, tokenAddress common.Address) (*PreTransaction, error) {
	nonce, err := c.client.PendingNonceAt(context.Background(), accountAddress)
	if err != nil {
		return nil, err
	}
	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	gasLimit, err := c.client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: tokenAddress,
	})
	if err != nil {
		return nil, err
	}
	return &PreTransaction{
		nonce:    nonce,
		gasPrice: gasPrice,
		gasLimit: gasLimit + (gasLimit * 20 / 100),
	}, nil
}

func (c *Client) Transfer(from *wallet.Account, to common.Address, amount *big.Float) (*types.Transaction, error) {
	b, err := c.GetBalance(from.Address())
	if err != nil {
		return nil, err
	}
	balance := BigIntToBigFloat(b, 18)
	if balance.Cmp(amount) < 0 {
		return nil, errors.New("insufficient balance")
	}
	amountWei := BigFloatToBigInt(amount, 18)
	t, err := c.GetPreTransaction(from.Address(), common.Address{})
	if err != nil {
		return nil, err
	}
	tx := types.NewTransaction(t.nonce, from.Address(), amountWei, t.gasLimit, t.gasPrice, []byte{})
	signedTx, err := c.SignTransaction(from, tx)
	if err != nil {
		return nil, err
	}
	err = c.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}
	return signedTx, nil
}

func (c *Client) SignTransaction(account *wallet.Account, tx *types.Transaction) (*types.Transaction, error) {
	chainID, err := c.client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}
	privateKey := account.Private()

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, err
	}
	return signedTx, nil
}
