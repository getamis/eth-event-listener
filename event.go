package listener

import (
	"github.com/ethereum/go-ethereum/core/types"
)

type BlockEvent struct {
	Block  *types.Block
	Events []*ContractEvent
}

type ContractEvent struct {
	Contract *Contract
	Tx       *types.Transaction
	Name     string
	Data     []byte
}

// func (c ContractEvent) Unpack(output interface{}) error {
// 	return c.Contract.abi.Unpack(output, c.Name, c.Data)
// }
