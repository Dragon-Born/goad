package wallet

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

type Account struct {
	privateKey   *ecdsa.PrivateKey
	tokenBalance map[string]*big.Int
	BNBBalance   *big.Int
}

func (a *Account) TokenBalance(symbol string) (balance *big.Int) {
	balance, ok := a.tokenBalance[symbol]
	if !ok {
		balance = big.NewInt(0)
	}
	return
}

func (a *Account) SetTokenBalance(symbol string, balance *big.Int) bool {
	if a.tokenBalance[symbol] == balance {
		return false
	}
	a.tokenBalance[symbol] = balance
	return true
}

func toECDSAFromHex(hexString string) (*ecdsa.PrivateKey, error) {
	pk := new(ecdsa.PrivateKey)
	var ok bool
	pk.D, ok = new(big.Int).SetString(hexString, 16)
	if !ok {
		return nil, errors.New("invalid private key")
	}
	pk.PublicKey.Curve = secp256k1.S256()
	pk.PublicKey.X, pk.PublicKey.Y = pk.PublicKey.Curve.ScalarBaseMult(pk.D.Bytes())
	return pk, nil
}

func IsValidBSCAddress(address string) bool {
	// IsHexAddress checks if a string can represent a valid hex-encoded
	// Ethereum address. Note that it doesn't check if the address has the
	// valid checksum capitalization.
	return common.IsHexAddress(address)
}

func FromPrivateKey(privateKey string) (*Account, error) {
	privateKey = strings.TrimLeft(privateKey, "0x")
	if len(privateKey) != 64 {
		return nil, errors.New("invalid private key length")
	}
	private, err := toECDSAFromHex(privateKey)
	if err != nil {
		return nil, err
	}
	return &Account{privateKey: private, tokenBalance: make(map[string]*big.Int)}, nil
}

func FromMnemonic(mnemonic string) (*Account, error) {
	wallet, err := NewHDWallet(mnemonic)
	if err != nil {
		return nil, err
	}
	return wallet.GetAccount(44, 195, 0, 0, 0)
}

func (a *Account) Random() {
	a.privateKey, _ = crypto.GenerateKey()
}

func (a *Account) Address() common.Address {
	return crypto.PubkeyToAddress(*a.privateKey.Public().(*ecdsa.PublicKey))
}

func (a *Account) AddressMask(noColor ...bool) string {

	maskAddress := crypto.PubkeyToAddress(*a.privateKey.Public().(*ecdsa.PublicKey)).String()
	if noColor != nil && noColor[0] {
		return fmt.Sprintf("%s...%s", maskAddress[:6], maskAddress[len(maskAddress)-4:])
	}
	return color.CyanString(fmt.Sprintf("%s...%s", maskAddress[:6], maskAddress[len(maskAddress)-4:]))
}

func (a *Account) Private() *ecdsa.PrivateKey {
	return a.privateKey
}
