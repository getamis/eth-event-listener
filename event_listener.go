package listener

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

const (
	bufferedLogSize = 1000
)

var logger = log.New()

type EventListener struct {
	client EthClient
	logCh  chan types.Log

	// Contract address <-> Contract mapping
	addressMap map[common.Address]*Contract
}

func NewEventListener(client EthClient,
	contracts []*Contract) *EventListener {

	l := &EventListener{
		client:     client,
		addressMap: make(map[common.Address]*Contract),
		logCh:      make(chan types.Log, bufferedLogSize),
	}

	for _, c := range contracts {
		l.addressMap[c.Address] = c
	}

	return l
}

func (el *EventListener) Listen(fromBlock *big.Int, eventCh chan<- *ContractEvent, stop <-chan struct{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	contracts := make([]common.Address, 0)
	for addr, _ := range el.addressMap {
		contracts = append(contracts, addr)
	}
	q := ethereum.FilterQuery{
		FromBlock: fromBlock,
		Addresses: contracts,
	}
	sub, err := el.client.SubscribeFilterLogs(ctx, q, el.logCh)
	if err != nil {
		logger.Error("Failed to subscribe filter logs", "err", err)
		return err
	}
	defer sub.Unsubscribe()

	// fetch the past logs
	logs, err := el.client.FilterLogs(context.Background(), q)
	if err != nil {
		logger.Error("Failed to execute a filter query command", "err", err)
		return err
	}

	go func() {
		for _, l := range logs {
			el.logCh <- l
		}
	}()

	for {
		select {
		case err := <-sub.Err():
			logger.Error("Unexpected subscription error", "err", err)
			return err
		case log := <-el.logCh:
			if cEvent := el.Parse(log); cEvent != nil {
				eventCh <- cEvent
			}
		case <-stop:
			logger.Warn("Received a stop signal")
			return nil
		}
	}
}

func (el *EventListener) Parse(l types.Log) *ContractEvent {
	c, ok := el.addressMap[l.Address]
	if !ok {
		return nil
	}
	if len(l.Topics) == 0 {
		return nil
	}
	name, ok := c.events[l.Topics[0]]
	if !ok {
		return nil
	}
	return &ContractEvent{
		BlockNumber: l.BlockNumber,
		BlockHash:   l.BlockHash,
		TxHash:      l.TxHash,
		Contract:    c,
		Name:        name,
		Data:        l.Data,
		Removed:     l.Removed,
	}
}
