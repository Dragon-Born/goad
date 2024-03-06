package blockchain

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"regexp"
	"strings"
)

func GetSupportedList() map[string]string {
	functionSignatures := []string{
		"swapExactTokensForTokensSupportingFeeOnTransferTokens(uint256 amountIn, uint256 amountOutMin, address[] path, address to, uint256 deadline)", // passed
		"swapExactTokensForTokensSupportingFeeOnTransferTokens(uint256 amountIn, uint256 amountOutMin, address[] path, address to)",                   // passed
		"swapExactETHForTokensSupportingFeeOnTransferTokens(uint256 amountIn, uint256 amountOutMin, address[] path, address to, uint256 deadline)",
		"swapExactETHForTokensSupportingFeeOnTransferTokens(uint256 amountIn, uint256 amountOutMin, address[] path, address to)",
		"swapExactTokensForETHSupportingFeeOnTransferTokens(uint256 amountIn, uint256 amountOutMin, address[] path, address to, uint256 deadline)", // passed
		"swapExactTokensForETHSupportingFeeOnTransferTokens(uint256 amountIn, uint256 amountOutMin, address[] path, address to)",                   // passed
		"swapTokensForExactTokens(uint256 amountIn, uint256 amountOutMin, address[] path, address to, uint256 deadline)",
		"swapTokensForExactTokens(uint256 amountIn, uint256 amountOutMin, address[] path, address to)",
		"swapExactETHForTokens(uint amountOutMin, address[] path, address to, uint deadline)",
		"swapExactETHForTokens(uint amountOutMin, address[] path, address to)",
		"swapTokensForExactETH(uint256 amountIn, uint256 amountOutMin, address[] path, address to, uint256 deadline)",
		"swapTokensForExactETH(uint256 amountIn, uint256 amountOutMin, address[] path, address to)",
		"swapExactTokensForETH(uint256 amountIn, uint256 amountOutMin, address[] path, address to, uint256 deadline)",
		"swapExactTokensForETH(uint256 amountIn, uint256 amountOutMin, address[] path, address to)",
		"swapETHForExactTokens(uint amountOut, address[] path, address to, uint deadline)",
		"swapETHForExactTokens(uint amountOut, address[] path, address to)",
		"swapExactTokensForTokens(uint256 amountIn, uint256 amountOutMin, address[] path, address to, uint256 deadline)",
		"swapExactTokensForTokens(uint256 amountIn, uint256 amountOutMin, address[] path, address to)",
		"multicall(uint256 deadline,bytes[] data)",
	}
	methods := make(map[string]string)
	for _, signature := range functionSignatures {
		re := regexp.MustCompile(`^(\w+)\((.*)\)$`)
		matches := re.FindStringSubmatch(signature)
		if matches == nil || len(matches) != 3 {
			log.Fatalf("Failed to parse the function signature")
		}
		methodName := matches[1]
		paramString := strings.ReplaceAll(matches[2], ", ", ",")
		paramTypes := strings.Split(paramString, ",")
		for i, paramType := range paramTypes {
			paramTypes[i] = strings.Split(paramType, " ")[0]
		}
		x := fmt.Sprintf("%s(%s)", methodName, strings.Join(paramTypes, ","))
		hash := crypto.Keccak256Hash([]byte(x))
		methodID := hash.Hex()[:10]
		methods[methodID] = signature
	}
	return methods
}
