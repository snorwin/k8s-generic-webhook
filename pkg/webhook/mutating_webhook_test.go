package webhook_test

import (
	"context"
	"encoding/json"
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
			result := (&webhook.MutateFunc{}).Mutate(context.TODO(), admission.Request{}, nil)
			Ω(result.Allowed).Should(BeTrue())
		})
		It("should use defined functions", func() {
			result := (&webhook.MutateFunc{
				Func: func(ctx context.Context, _ admission.Request, _ runtime.Object) admission.Response {
					return admission.Denied("")
				},
			}).Mutate(context.TODO(), admission.Request{}, nil)
			Ω(result.Allowed).Should(BeFalse())
		})
	})
	Context("PatchResponseFromObject", func() {
		It("should not create patch if object was not modified", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
			}
			raw, err := json.Marshal(pod)
			Ω(err).ShouldNot(HaveOccurred())

			response := webhook.PatchResponseFromObject(admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: raw,
					},
					Operation: admissionv1.Create,
				},
			}, pod)

			Ω(response.Allowed).Should(BeTrue())
			Ω(response.Patches).Should(BeEmpty())
		})
		It("should create patches", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
			}
			raw, err := json.Marshal(pod)
			Ω(err).ShouldNot(HaveOccurred())

			modified := pod.DeepCopy()
			modified.Name = "bar"

			response := webhook.PatchResponseFromObject(admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: raw,
					},
					Operation: admissionv1.Create,
				},
			}, modified)

			Ω(response.Allowed).Should(BeTrue())
			Ω(response.Patches).ShouldNot(BeEmpty())
		})
	})
})
