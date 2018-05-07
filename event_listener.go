package listener

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

var logger = log.New()

type EventListener struct {
	client  EthClient
	logCh   chan types.Log
	eventCh chan *ContractEvent

	// Contract address <-> Contract mapping
	addressMap map[common.Address]*Contract
}

func NewEventListener(client EthClient,
	contracts []*Contract,
	bufferedLogSize int,
	bufferedEventSize int) *EventListener {

	l := &EventListener{
		client:     client,
		addressMap: make(map[common.Address]*Contract),
		logCh:      make(chan types.Log, bufferedLogSize),
		eventCh:    make(chan *ContractEvent, bufferedEventSize),
	}

	for _, c := range contracts {
		l.addressMap[c.Address] = c
	}

	return l
}

func (el *EventListener) Listen(fromBlock *big.Int, stop <-chan struct{}) error {
	wg := sync.WaitGroup{}
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
		return err
	}

	// fetch the past logs
	logs, err := el.client.FilterLogs(context.Background(), q)
	if err != nil {
		return err
	}
	defer el.channelCleanUp()
	defer sub.Unsubscribe()

	wg.Add(1)
	defer wg.Wait()
	go func() {
		for _, l := range logs {
			el.logCh <- l
		}
		wg.Done()
	}()

	for {
		select {
		case err := <-sub.Err():
			return err
		case log := <-el.logCh:
			if cEvent := el.Parse(log); cEvent != nil {
				el.eventCh <- cEvent
			}
		case <-stop:
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

func (el *EventListener) channelCleanUp() {
	// Unsubscribe should be called before this cleanUp stage, therefore geth
	// would stop sending logs through the log channel (but it won't close it).
	// The goal of this function is to drain the log channel, send events through
	// event channel and close it (to notify receiver that there's no more data).
	for len(el.logCh) > 0 {
		log := <-el.logCh
		if cEvent := el.Parse(log); cEvent != nil {
			el.eventCh <- cEvent
		}
	}
	close(el.eventCh)
	return
}
