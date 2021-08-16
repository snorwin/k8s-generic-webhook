package webhook_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/snorwin/k8s-generic-webhook/pkg/webhook"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ = Describe("Simple Mutating Webhook", func() {
	Context("SimpleMutateFunc", func() {
		It("should by default allow all", func() {
			result := (&webhook.SimpleMutateFunc{}).Mutate(context.TODO(), admission.Request{}, nil)
			Ω(result).Should(BeNil())
		})
		It("should use defined functions", func() {
			result := (&webhook.SimpleMutateFunc{
				Func: func(ctx context.Context, _ admission.Request, _ runtime.Object) error {
					return errors.New("error")
				},
			}).Mutate(context.TODO(), admission.Request{}, nil)
			Ω(result).Should(Equal(errors.New("error")))
		})
	})
})
