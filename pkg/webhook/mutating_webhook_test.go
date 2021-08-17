package webhook_test

import (
	"context"
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/snorwin/k8s-generic-webhook/pkg/webhook"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	Context("MutateObjectFunc", func() {
		var (
			n   *corev1.Namespace
			raw []byte
		)
		BeforeEach(func() {
			var err error
			n = &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			}
			raw, err = json.Marshal(n)
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should by default allow all", func() {
			result := (&webhook.MutateObjectFunc{}).Mutate(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeTrue())
		})
		It("should use defined functions", func() {
			result := (&webhook.MutateObjectFunc{
				Func: func(ctx context.Context, _ admission.Request, object runtime.Object) error {
					Ω(object).Should(Equal(n))
					return nil
				},
			}).Mutate(context.TODO(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Object: runtime.RawExtension{
						Object: n,
						Raw:    raw,
					},
				},
			})
			Ω(result.Allowed).Should(BeTrue())
		})
		It("should deny if error", func() {
			result := (&webhook.MutateObjectFunc{
				Func: func(ctx context.Context, _ admission.Request, _ runtime.Object) error {
					return errors.New("")
				},
			}).Mutate(context.TODO(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Object: runtime.RawExtension{
						Object: n,
						Raw:    raw,
					},
				},
			})
			Ω(result.Allowed).Should(BeFalse())
		})
	})
})
