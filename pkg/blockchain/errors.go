package blockchain

import (
	"errors"
)

var (
	ErrorBCNotFound = errors.New("blockchain not found")
	ErrorBCExists   = errors.New("blockchain already exists")

	ErrorTxNotFound     = errors.New("transaction not found")
	ErrorTxSignFailed   = errors.New("transaction signing failed")
	ErrorTxCreateFailed = errors.New("transaction creation failed")
)
