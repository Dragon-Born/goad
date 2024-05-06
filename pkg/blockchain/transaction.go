package blockchain

import (
	"encoding/hex"
	"errors"
	commonSol "github.com/blocto/solana-go-sdk/common"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"math/big"
	"regexp"
	"strings"
)

type DexTransaction struct {
	From     common.Address
	Method   string
	Amount   *big.Int
	GasPrice *big.Int
}

type SolDexTransaction struct {
	From     commonSol.PublicKey
	Method   string
	Amount   *big.Int
	GasPrice *big.Int
}

func DecodeTransaction(txData string) (args map[string]any, err error) {
	args = make(map[string]interface{})
	re := regexp.MustCompile(`^(\w+)\((.*)\)$`)
	signature := GetSupportedList()[txData[:10]]
	matches := re.FindStringSubmatch(signature)
	if matches == nil || len(matches) != 3 {
		//log.Error("Failed to parse the function signature")
		err = errors.New("failed to parse the function signature")
		return
	}
	methodName := matches[1]
	paramString := strings.ReplaceAll(matches[2], ", ", ",")
	paramTypes := strings.Split(paramString, ",")
	var inputs []abi.Argument
	count := 0
	for _, paramType := range paramTypes {
		splitTypes := strings.Split(paramType, " ")
		abiType := parseParamType(splitTypes[0])
		inputs = append(inputs, abi.Argument{
			Name: splitTypes[1],
			Type: abiType,
		})
		count++
	}
	var outputs []abi.Argument
	method := abi.NewMethod(methodName, methodName, abi.Function, "nonpayable", false, false, inputs, outputs)
	contractABI := abi.ABI{
		Methods: map[string]abi.Method{
			methodName: method,
		},
	}

	data, err := hex.DecodeString(txData[2:])
	if err != nil {
		//log.Errorf("Failed to decode tx input: %v", err)
		return
	}
	methodID := data[:4]
	data = data[4:]

	m, err := contractABI.MethodById(methodID)
	if err != nil {
		//log.Errorf("Method not found in ABI: %v", err)
		return
	}
	err = m.Inputs.UnpackIntoMap(args, data)
	if err != nil {
		//log.Errorf("Failed to unpack method inputs: %v", err)
		return
	}
	return
}
