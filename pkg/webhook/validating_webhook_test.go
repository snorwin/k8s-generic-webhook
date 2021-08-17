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

var _ = Describe("Validating Webhook", func() {
	Context("ValidateFuncs", func() {
		It("should by default allow all", func() {
			result := (&webhook.ValidateFuncs{}).ValidateCreate(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeTrue())
			result = (&webhook.ValidateFuncs{}).ValidateUpdate(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeTrue())
			result = (&webhook.ValidateFuncs{}).ValidateDelete(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeTrue())
		})
		It("should use defined functions", func() {
			result := (&webhook.ValidateFuncs{
				CreateFunc: func(ctx context.Context, _ admission.Request) admission.Response {
					return admission.Denied("")
				},
			}).ValidateCreate(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeFalse())
		})
		It("should by default allow all", func() {
			result := (&webhook.ValidateObjectFuncs{}).ValidateCreate(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeTrue())
			result = (&webhook.ValidateObjectFuncs{}).ValidateUpdate(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeTrue())
			result = (&webhook.ValidateObjectFuncs{}).ValidateDelete(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeTrue())
		})
		It("should use defined functions", func() {
			result := (&webhook.ValidateObjectFuncs{
				CreateFunc: func(ctx context.Context, _ admission.Request, _ runtime.Object) error {
					return errors.New("")
				},
			}).ValidateCreate(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeFalse())
			result = (&webhook.ValidateObjectFuncs{
				CreateFunc: func(ctx context.Context, _ admission.Request, _ runtime.Object) error {
					return nil
				},
			}).ValidateCreate(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeTrue())

			result = (&webhook.ValidateObjectFuncs{
				UpdateFunc: func(ctx context.Context, _ admission.Request, _ runtime.Object, _ runtime.Object) error {
					return errors.New("")
				},
			}).ValidateUpdate(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeFalse())
			result = (&webhook.ValidateObjectFuncs{
				UpdateFunc: func(ctx context.Context, _ admission.Request, _ runtime.Object, _ runtime.Object) error {
					return nil
				},
			}).ValidateUpdate(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeTrue())

			result = (&webhook.ValidateObjectFuncs{
				DeleteFunc: func(ctx context.Context, _ admission.Request, _ runtime.Object) error {
					return errors.New("")
				},
			}).ValidateDelete(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeFalse())
			result = (&webhook.ValidateObjectFuncs{
				DeleteFunc: func(ctx context.Context, _ admission.Request, _ runtime.Object) error {
					return nil
				},
			}).ValidateDelete(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeTrue())
		})

	})
})
