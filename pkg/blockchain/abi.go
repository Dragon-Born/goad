package blockchain

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	log "github.com/sirupsen/logrus"
)

func parseParamType(param string) abi.Type {
	t, err := abi.NewType(param, "", nil)
	if err != nil {
		log.Error("Failed to create ABI type for param: %s, error: %v", param, err)
	}
	return t
}
