package wallet

import (
	"fmt"
	"github.com/miguelmota/go-ethereum-hdwallet"
)

type HDWallet struct {
	wallet *hdwallet.Wallet
}

func NewHDWallet(mnemonic string) (*HDWallet, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}
	return &HDWallet{
		wallet: wallet,
	}, nil
}

func (w *HDWallet) GetAccount(purpose, coinType, account, change, index uint32) (*Account, error) {
	pathStr := fmt.Sprintf("m/%d'/%d'/%d'/%d/%d", purpose, coinType, account, change, index)
	path := hdwallet.MustParseDerivationPath(pathStr)
	acc, err := w.wallet.Derive(path, false)
	if err != nil {
		return nil, err
	}
	private, err := w.wallet.PrivateKey(acc)
	if err != nil {
		return nil, err
	}
	return &Account{privateKey: private}, nil
}
