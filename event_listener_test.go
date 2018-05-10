package listener

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/stretchr/testify/mock"

	"github.com/getamis/eth-event-listener/mocks"
)

var _ = Describe("Event listener tests", func() {
	var (
		mockClient *mocks.EthClient
		l          *EventListener
		stop       chan struct{}
	)
	testEventID := hashGen()
	testEvents := make(map[common.Hash]string)
	testEvents[testEventID] = "test-event"
	testContracts := []*Contract{
		&Contract{
			Name:    "test-contract",
			Address: addrGen(),
			events:  testEvents,
		},
	}

	BeforeEach(func() {
		mockClient = &mocks.EthClient{}
		l = NewEventListener(mockClient, testContracts, 10000, 10000)
		stop = make(chan struct{}, 1)
	})

	Context("Listen tests", func() {
		It("SubscribeFilterLogs failed", func() {
			expectedErr := errors.New("SubscribeFilterLogs failed")
			mockClient.On("SubscribeFilterLogs", Anything, Anything, Anything).Return(nil, expectedErr).Once()
			err := l.Listen(nil, stop)
			Expect(expectedErr).Should(Equal(err))
		})

		It("FilterLogs failed", func() {
			emptySub := &Subscription{
				err: make(chan error, 1),
			}
			mockClient.On("SubscribeFilterLogs", Anything, Anything, Anything).Return(emptySub, nil).Once()
			expectedErr := errors.New("FilterLogs failed")
			mockClient.On("FilterLogs", Anything, Anything).Return(nil, expectedErr).Once()
			err := l.Listen(nil, stop)
			Expect(expectedErr).Should(Equal(err))
		})

		It("Subscription Error", func() {
			expectedErr := errors.New("Subscription error")
			errCh := make(chan error, 1)
			emptySub := &Subscription{
				err: errCh,
			}

			go func() {
				errCh <- expectedErr
			}()
			mockClient.On("SubscribeFilterLogs",
				Anything, Anything, Anything).Return(emptySub, nil)
			mockClient.On("FilterLogs", Anything, Anything).Return(nil, nil).Once()
			err := l.Listen(nil, stop)
			Expect(expectedErr).Should(Equal(err))
		})

		It("Handle the past log", func() {
			errCh := make(chan error, 1)
			emptySub := &Subscription{
				err: errCh,
			}
			mockClient.On("SubscribeFilterLogs",
				Anything, Anything, Anything).Return(emptySub, nil)

			blockNumber := uint64(1)
			blockHash := hashGen()
			txHash := hashGen()
			pastLog := types.Log{
				Address:     testContracts[0].Address,
				Topics:      []common.Hash{testEventID},
				BlockNumber: blockNumber,
				BlockHash:   blockHash,
				TxHash:      txHash,
			}
			mockClient.On("FilterLogs", Anything, Anything).Return([]types.Log{pastLog}, nil).Once()

			go l.Listen(nil, stop)

			eventCh := l.GetEventCh()
			var event *ContractEvent = nil
			select {
			case event = <-eventCh:
			case <-time.After(1 * time.Second):
			}

			expectedEvent := &ContractEvent{
				BlockNumber: blockNumber,
				BlockHash:   blockHash,
				TxHash:      txHash,
				Contract:    testContracts[0],
				Name:        testEvents[testEventID],
			}
			Expect(expectedEvent).Should(Equal(event))
			close(stop)
		})

		It("Gracefully shut down", func() {
			errCh := make(chan error, 1)
			emptySub := &Subscription{
				err: errCh,
			}
			mockClient.On("SubscribeFilterLogs",
				Anything, Anything, Anything).Return(emptySub, nil)

			var pastLogs []types.Log
			var receivedEvents []*ContractEvent
			pastLogNum := 9999
			for i := 0; i < pastLogNum; i++ {
				blockNumber := uint64(1)
				blockHash := hashGen()
				txHash := hashGen()
				log := types.Log{
					Address:     testContracts[0].Address,
					Topics:      []common.Hash{testEventID},
					BlockNumber: blockNumber,
					BlockHash:   blockHash,
					TxHash:      txHash,
				}
				pastLogs = append(pastLogs, log)
			}
			mockClient.On("FilterLogs", Anything, Anything).Return(pastLogs, nil).Once()
			go l.Listen(nil, stop)
			close(stop) //shut it down immediately
			time.Sleep(5 * time.Second)

			eventCh := l.GetEventCh()
			for event := range eventCh {
				receivedEvents = append(receivedEvents, event)
			}
			Expect(len(receivedEvents)).Should(Equal(pastLogNum))
		})
	})
})

func TestApplicationSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Event Listener Suite")
}

type Subscription struct {
	err chan error
}

func (s Subscription) Unsubscribe() {}

func (s Subscription) Err() <-chan error {
	return s.err
}

func hashGen() common.Hash {
	return common.StringToHash(fmt.Sprintf("hash-%d", time.Now().UnixNano()))
}

func addrGen() common.Address {
	return common.StringToAddress(fmt.Sprintf("addr-%d", time.Now().UnixNano()))
}
