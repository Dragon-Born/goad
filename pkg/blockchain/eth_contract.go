package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"goad/pkg/cache"
	"goad/pkg/wallet"
	"io/ioutil"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Contract struct {
	*Client
	address     common.Address
	name        string
	symbol      string
	decimals    int
	totalSupply int
}

func (c *Client) NewContract(address common.Address) *Contract {
	return &Contract{
		Client:  c,
		address: address,
	}
}

func (c *Contract) GetTokenBalance(accountAddress common.Address) (*big.Int, error) {
	data := MethodPack("balanceOf(address)", common.LeftPadBytes(accountAddress.Bytes(), 32))
	msg := ethereum.CallMsg{From: common.Address{}, To: &c.address, Data: data}
	result, err := c.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, err
	}
	//decimals, err := c.GetTokenDecimals()
	//if err != nil {
	//	return nil, err
	//}
	return common.BytesToHash(result[:]).Big(), nil
}

func (c *Contract) GetTokenDecimals() (int, error) {
	if c.decimals != 0 {
		return c.decimals, nil
	}
	data := MethodPack("decimals()")
	msg := ethereum.CallMsg{From: common.Address{}, To: &c.address, Data: data}
	result, err := c.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return 0, err
	}
	return int(common.BytesToHash(result[:]).Big().Int64()), nil
}

func (c *Contract) GetTokenSymbol() (string, error) {

	if c.symbol != "" {
		return strings.TrimSpace(c.symbol), nil
	}
	data := MethodPack("symbol()")
	msg := ethereum.CallMsg{From: common.Address{}, To: &c.address, Data: data}
	result, err := c.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return "", err
	}
	var nonAlphanumericRegex = regexp.MustCompile(`[^\p{L}\p{N}]+`)
	symbol := nonAlphanumericRegex.ReplaceAllString(string(result), "")

	c.symbol = symbol
	return c.symbol, nil
}

func (c *Contract) GetTokenName() (string, error) {
	if c.name != "" {
		return c.name, nil
	}
	data := MethodPack("name()")
	msg := ethereum.CallMsg{From: common.Address{}, To: &c.address, Data: data}
	result, err := c.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return "", err
	}
	var nonAlphanumericRegex = regexp.MustCompile(`[^\p{L}\p{N}]+`)
	name := nonAlphanumericRegex.ReplaceAllString(string(result), "")
	return name, nil
}

func (c *Contract) GetTokenTotalSupply() (*big.Float, error) {
	if c.totalSupply != 0 {
		return nil, nil
	}
	data := MethodPack("totalSupply()")
	msg := ethereum.CallMsg{From: common.Address{}, To: &c.address, Data: data}
	result, err := c.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, err
	}
	decimals, err := c.GetTokenDecimals()
	if err != nil {
		return nil, err
	}
	return BigIntToBigFloat(common.BytesToHash(result[:]).Big(), decimals), nil
}

func (c *Contract) GetTokenAllowance(ownerAddress, spenderAddress common.Address) (*big.Float, error) {
	data := MethodPack(
		"allowance(address,address)",
		common.LeftPadBytes(ownerAddress.Bytes(), 32),
		common.LeftPadBytes(spenderAddress.Bytes(), 32),
	)
	msg := ethereum.CallMsg{From: common.Address{}, To: &c.address, Data: data}
	result, err := c.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, err
	}
	decimals, err := c.GetTokenDecimals()
	if err != nil {
		return nil, err
	}
	return BigIntToBigFloat(common.BytesToHash(result[:]).Big(), decimals), nil
}

func (c *Contract) TransferToken(from *wallet.Account, toAddress common.Address, amount float64) (*types.Transaction, error) {
	decimals, err := c.GetTokenDecimals()
	if err != nil {
		return nil, err
	}
	amountWei := AmountToBigInt(amount, decimals)
	data := MethodPack(
		"transfer(address,uint256)",
		common.LeftPadBytes(toAddress.Bytes(), 32),
		common.LeftPadBytes(amountWei.Bytes(), 32),
	)
	t, err := c.GetPreTransaction(from.Address(), c.address)
	if err != nil {
		return nil, err
	}

	tx := types.NewTransaction(t.nonce, c.address, big.NewInt(0), t.gasLimit, t.gasPrice, data)
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

func (c *Contract) GetCode() ([]byte, error) {
	return c.client.CodeAt(context.Background(), c.address, nil)

}

func (c *Contract) GetPrice() (float64, error) {
	cIdentifier := fmt.Sprintf("%v_price", c.symbol)
	ca := cache.GetCache()
	nei, ok := ca.Get(cIdentifier)
	if ok {
		return nei.(float64), nil
	}
	symbol, err := c.GetTokenSymbol()
	if err != nil {
		return 0, err
	}
	price, err := TokenPrice(symbol)
	if err != nil {
		return 0, err
	}
	ca.Set(cIdentifier, price, time.Minute*60)
	return price, nil
}

func TokenPrice(tokenIdentifier string) (float64, error) {
	const apiURL = "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest"
	const apiKey = "bed5d884-3e5d-411c-8aa6-002cd8944de8"
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %v", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("symbol", tokenIdentifier) // or "address" for contract address
	req.URL.RawQuery = q.Encode()

	// Set API Key
	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// Check for API-specific errors
	if status, ok := result["status"].(map[string]interface{}); ok {
		if errorCode, ok := status["error_code"].(float64); ok && errorCode != 0 {
			errorMessage := status["error_message"].(string)
			return 0, fmt.Errorf("API error %d: %s", int(errorCode), errorMessage)
		}
	}

	// Extract price from result
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid response structure: missing 'data'")
	}

	tokenData, ok := data[tokenIdentifier].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("token '%s' not found in response", tokenIdentifier)
	}

	quote, ok := tokenData["quote"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid response structure: missing 'quote'")
	}

	usd, ok := quote["USD"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid response structure: missing 'USD'")
	}

	price, ok := usd["price"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid response structure: missing 'price'")
	}

	return price, nil
}
