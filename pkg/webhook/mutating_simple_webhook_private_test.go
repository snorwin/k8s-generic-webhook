package webhook

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ = Describe("Simple Mutating Webhook", func() {
	Context("wrapAsMutator", func() {
		It("should inject the decoder", func() {
			dec := &admission.Decoder{}
			sm := &SimpleMutatingWebhook{}
			m := wrapAsMutator(sm)

			sma, ok := m.(*simpleMutatorAdapter)
			Ω(ok).Should(BeTrue())
			err := sma.InjectDecoder(dec)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(sm.Decoder).Should(Equal(dec))
		})
		It("should inject the client", func() {
			cl := fake.NewClientBuilder().Build()
			sm := &SimpleMutatingWebhook{}
			m := wrapAsMutator(sm)

			sma, ok := m.(*simpleMutatorAdapter)
			Ω(ok).Should(BeTrue())
			err := sma.InjectClient(cl)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(sm.Client).Should(Equal(cl))
		})
	})
})
