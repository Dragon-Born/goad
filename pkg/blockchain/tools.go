package blockchain

import (
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/sha3"
)

func AmountToBigInt(val float64, decimals int) *big.Int {
	bigval := new(big.Float)
	bigval.SetFloat64(val)
	coin := new(big.Float)
	dec := new(big.Int)
	dec.SetInt64(int64(decimals))
	coin.SetInt(big.NewInt(10).Exp(big.NewInt(10), dec, nil))
	bigval.Mul(bigval, coin)
	result := new(big.Int)
	bigval.Int(result)
	return result
}

func BigIntToAmount(val *big.Int, decimals int) float64 {
	bigval := new(big.Float)
	bigval.SetInt(val)
	coin := new(big.Float)
	dec := new(big.Int)
	dec.SetInt64(int64(decimals))
	coin.SetInt(big.NewInt(10).Exp(big.NewInt(10), dec, nil))
	bigval.Quo(bigval, coin)
	result, _ := bigval.Float64()
	return result
}

func Int64ToAmount(val int64, decimals int) float64 {
	bigval := new(big.Float)
	bigval.SetInt64(val)

	coin := new(big.Float)
	dec := new(big.Int)
	dec.SetInt64(int64(decimals))
	coin.SetInt(big.NewInt(10).Exp(big.NewInt(10), dec, nil))
	bigval.Quo(bigval, coin)
	result, _ := bigval.Float64()
	return result
}

func BigIntToBigFloat(val *big.Int, decimals int) *big.Float {
	var prec uint = 256
	bigval := new(big.Float).SetPrec(prec)
	bigval.SetInt(val)
	coin := new(big.Float).SetPrec(prec)
	dec := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	coin.SetInt(dec)
	bigval.Quo(bigval, coin)
	return bigval
}

func BigFloatToBigInt(val *big.Float, decimals int) *big.Int {
	bigval := new(big.Float)
	bigval.Set(val)
	coin := new(big.Float)
	dec := new(big.Int)
	dec.SetInt64(int64(decimals))
	coin.SetInt(big.NewInt(10).Exp(big.NewInt(10), dec, nil))
	bigval.Mul(bigval, coin)
	result := new(big.Int)
	bigval.Int(result)
	return result
}
func MethodPack(method string, args ...[]byte) []byte {
	fnSignature := []byte(method)
	hash := sha3.NewLegacyKeccak256()
	hash.Write(fnSignature)
	methodID := hash.Sum(nil)[:4]
	var data []byte
	data = append(data, methodID...)
	for _, arg := range args {
		data = append(data, arg...)
	}
	return data
}

func BigFloatToString(amount *big.Float) string {
	str := fmt.Sprintf("%.2f", amount)
	splitStr := strings.Split(str, ".")
	splitLengh := len(splitStr[1]) + 1
	dotIndex := len(str) - splitLengh
	integerPart := str[:dotIndex]
	decimalPart := str[dotIndex:]
	for i := dotIndex - 3; i > 0; i -= 3 {
		integerPart = integerPart[:i] + "," + integerPart[i:]
	}
	return integerPart + decimalPart
}
