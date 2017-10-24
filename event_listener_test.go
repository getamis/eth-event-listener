package listener

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/maichain/listener/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/stretchr/testify/mock"
)

var _ = Describe("Event listener tests", func() {
	var (
		l          *EventListener
		mockClient *mocks.EthClient
	)

	BeforeEach(func() {
		mockClient = &mocks.EthClient{}
		l = NewEventListener(mockClient, nil)
	})

	Context("Listen tests", func() {
		It("SubscribeNewHead failed", func() {
			expectedErr := errors.New("SubscribeNewHead failed")
			mockClient.On("SubscribeNewHead", Anything, Anything).Return(nil, expectedErr)
			err := l.Listen(nil)
			Expect(expectedErr).Should(Equal(err))
		})

		It("Subscribe failed", func() {
			expectedErr := errors.New("Subscribe failed")
			errCh := make(chan error, 1)
			emptySub := &Subscription{
				err: errCh,
			}

			go func() {
				errCh <- expectedErr
			}()
			mockClient.On("SubscribeNewHead",
				Anything, Anything).Return(emptySub, nil)
			err := l.Listen(nil)
			Expect(expectedErr).Should(Equal(err))
		})

		It("New block comes", func() {
			errCh := make(chan error, 1)
			blockEventCh := make(chan *BlockEvent, 1)
			emptySub := &Subscription{
				err: errCh,
			}
			expectedHeader := &types.Header{}
			expectedBlock := types.NewBlockWithHeader(expectedHeader)

			mockClient.On("SubscribeNewHead", Anything, Anything).Return(emptySub, nil)
			mockClient.On("BlockByHash", Anything, expectedHeader.Hash()).Return(expectedBlock, nil)
			go func() {
				l.subCh <- expectedHeader
			}()

			go l.Listen(blockEventCh)
			blockEvent := <-blockEventCh
			expectedBlockEvent := &BlockEvent{
				Block: expectedBlock,
			}
			Expect(expectedBlockEvent).Should(Equal(blockEvent))
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
