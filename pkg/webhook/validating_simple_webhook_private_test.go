package webhook

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ = Describe("Simple Validating Webhook", func() {
	Context("wrapAsMutator", func() {
		It("should inject the decoder", func() {
			dec := &admission.Decoder{}
			sv := &SimpleValidatingWebhook{}
			v := wrapAsValidator(sv)

			sva, ok := v.(*simpleValidatorAdapter)
			Ω(ok).Should(BeTrue())
			err := sva.InjectDecoder(dec)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(sv.Decoder).Should(Equal(dec))
		})
		It("should inject the client", func() {
			cl := fake.NewClientBuilder().Build()
			sv := &SimpleValidatingWebhook{}
			v := wrapAsValidator(sv)

			sva, ok := v.(*simpleValidatorAdapter)
			Ω(ok).Should(BeTrue())
			err := sva.InjectClient(cl)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(sv.Client).Should(Equal(cl))
		})
	})
})
