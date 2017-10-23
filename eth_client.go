package listener

import (
	"context"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type EthClient interface {
	// SubscribeNewHead subscribes to notifications about the current blockchain head
	// on the given channel.
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
	// TransactionReceipt returns the receipt of a transaction by transaction hash.
	// Note that the receipt is not available for pending transactions.
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	// BlockByHash returns the given full block.
	//
	// Note that loading full blocks requires two requests. Use HeaderByHash
	// if you don't need all transactions or uncle headers.
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
}
