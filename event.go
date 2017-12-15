package listener

import (
	"github.com/ethereum/go-ethereum/common"
)

type ContractEvent struct {
	// block in which the transaction was included
	BlockNumber uint64
	// hash of the block in which the transaction was included
	BlockHash common.Hash
	// hash of the transaction
	TxHash common.Hash

	// contract which the event belongs to
	Contract *Contract
	// name of the contract event
	Name string
	// supplied by the contract, usually ABI-encoded
	Data []byte

	// The Removed field is true if this log was reverted due to a chain reorganisation.
	// You must pay attention to this field if you receive logs through a filter query.
	Removed bool
}

//func (c ContractEvent) Unpack(output interface{}) error {
//	return c.Contract.abi.Unpack(output, c.Name, c.Data)
//}
