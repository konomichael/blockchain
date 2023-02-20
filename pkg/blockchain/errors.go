package blockchain

import (
	"errors"
)

var (
	ErrorBCNotFound = errors.New("blockchain not found")
	ErrorBCExists   = errors.New("blockchain already exists")

	ErrorBlkHeightInvalid   = errors.New("block height is invalid")
	ErrorBlkPrevHashInvalid = errors.New("block previous hash is invalid")

	ErrorTxNotFound     = errors.New("transaction not found")
	ErrorTxSignFailed   = errors.New("transaction signing failed")
	ErrorTxCreateFailed = errors.New("transaction creation failed")
	ErrorTxInvalid      = errors.New("transaction is invalid")
)
