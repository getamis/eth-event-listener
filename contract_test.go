package listener

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ethereum/go-ethereum/common"
)

var _ = Describe("Contract tests", func() {
	Context("NewContract tests", func() {
		It("valid abi file", func() {
			const name = "test_contract"
			const abi = `
			[
				{ "type" : "function", "name" : "send", "constant" : false, "inputs" : [ { "name" : "amount", "type" : "uint256" } ] },
				{ "type" : "event", "name" : "balance" }
			]`
			addr := common.StringToAddress("0x1234567890")

			c, err := NewContract(name, abi, addr)

			Expect(err).Should(BeNil())
			Expect(name).Should(Equal(c.Name))
			Expect(addr).Should(Equal(c.Address))
			Expect(c.ABI).ShouldNot(BeNil())
			Expect(len(c.events)).Should(BeNumerically("==", 1))
			for _, event := range c.ABI.Events {
				Expect(c.events[event.Id()]).ShouldNot(BeNil())
			}
		})

		It("invalid abi file", func() {
			const name = "test_contract"
			const abi = `wrong abi files`
			addr := common.StringToAddress("0x1234567890")

			_, err := NewContract(name, abi, addr)
			Expect(err).ShouldNot(BeNil())
		})
	})
})
