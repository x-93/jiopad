package testapi

import (
	"github.com/karlsen-network/karlsend/v2/domain/consensus/model"
	"github.com/karlsen-network/karlsend/v2/domain/consensus/utils/txscript"
)

// TestTransactionValidator adds to the main TransactionValidator methods required by tests
type TestTransactionValidator interface {
	model.TransactionValidator
	SigCache() *txscript.SigCache
	SetSigCache(sigCache *txscript.SigCache)
}
