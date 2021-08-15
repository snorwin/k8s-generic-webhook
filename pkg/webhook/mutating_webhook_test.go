package webhook_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/snorwin/k8s-generic-webhook/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ = Describe("Mutating Webhook", func() {
	Context("MutateFunc", func() {
		It("should by default allow all", func() {
			result := (&webhook.MutateFunc{}).Mutate(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeTrue())
		})
		It("should use defined functions", func() {
			result := (&webhook.MutateFunc{
				Func: func(ctx context.Context, _ admission.Request) admission.Response {
					return admission.Denied("")
				},
			}).Mutate(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeFalse())
		})
	})
})
