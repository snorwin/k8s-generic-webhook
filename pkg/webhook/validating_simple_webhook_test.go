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

var _ = Describe("Simple Validating Webhook", func() {
	Context("SimpleValidateFuncs", func() {
		It("should by default allow all", func() {
			result := (&webhook.SimpleValidateFuncs{}).ValidateCreate(context.TODO(), admission.Request{}, nil)
			Ω(result).Should(BeNil())
			result = (&webhook.SimpleValidateFuncs{}).ValidateUpdate(context.TODO(), admission.Request{}, nil, nil)
			Ω(result).Should(BeNil())
			result = (&webhook.SimpleValidateFuncs{}).ValidateDelete(context.TODO(), admission.Request{}, nil)
			Ω(result).Should(BeNil())
		})
		It("should use defined functions", func() {
			result := (&webhook.SimpleValidateFuncs{
				CreateFunc: func(ctx context.Context, _ admission.Request, _ runtime.Object) error {
					return errors.New("error")
				},
			}).ValidateCreate(context.TODO(), admission.Request{}, nil)
			Ω(result).Should(Equal(errors.New("error")))
		})
	})
})
