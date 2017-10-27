package listener

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/maichain/go-ethereum/log"
)

var logger = log.New()

type EventListener struct {
	client EthClient
	subCh  chan *types.Header

	// Contract name <-> Contract mapping
	nameMap map[string]*Contract
	// Contract address <-> Contract mapping
	addressMap map[common.Address]*Contract
}

func NewEventListener(client EthClient,
	contracts []*Contract) *EventListener {

	l := &EventListener{
		client:     client,
		nameMap:    make(map[string]*Contract),
		addressMap: make(map[common.Address]*Contract),
		subCh:      make(chan *types.Header),
	}

	for _, c := range contracts {
		l.nameMap[c.Name] = c
		l.addressMap[c.Address] = c
	}

	return l
}

func (el *EventListener) Listen(eventCh chan<- *BlockEvent, stop <-chan struct{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sub, err := el.client.SubscribeNewHead(ctx, el.subCh)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	for {
		select {
		case err := <-sub.Err():
			return err
		case header := <-el.subCh:
			block, err := el.client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				continue
			}

			eventCh <- el.Parse(block)
		case <-stop:
			return nil
		}
	}
}

func (el *EventListener) Parse(block *types.Block) *BlockEvent {
	blockEvent := &BlockEvent{
		Block: block,
	}
	for _, tx := range block.Transactions() {
		if tx.To() == nil {
			continue
		}

		// check if this event is not in our registered contracts
		c, ok := el.addressMap[*tx.To()]
		if !ok {
			continue
		}
		r, err := el.client.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			logger.Warn("Failed to get receipt", "err", err)
			continue
		}
		for _, l := range r.Logs {
			if len(l.Topics) == 0 {
				break
			}
			name, ok := c.events[l.Topics[0]]
			if ok {
				blockEvent.Events = append(blockEvent.Events, &ContractEvent{
					Contract: c,
					Tx:       tx,
					Name:     name,
					Data:     l.Data,
				})
			}
		}
	}
	return blockEvent
}
