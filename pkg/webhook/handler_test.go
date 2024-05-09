package webhook

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gomodules.xyz/jsonpatch/v2"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var _ = Describe("Handler", func() {
	Context("Handle", func() {
		var (
			decoder admission.Decoder
		)
		BeforeEach(func() {
			scheme := runtime.NewScheme()
			err := corev1.AddToScheme(scheme)
			Ω(err).ShouldNot(HaveOccurred())
			decoder = admission.NewDecoder(scheme)
			Ω(err).ShouldNot(HaveOccurred())

		})
		It("should deny by default", func() {
			result := (&handler{}).Handle(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeFalse())
		})
		It("should mutate and generate patches", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
			}
			raw, err := json.Marshal(pod)
			Ω(err).ShouldNot(HaveOccurred())

			h := withMutationHandler(&MutateFunc{
				Func: func(_ context.Context, _ admission.Request, obj runtime.Object) admission.Response {
					pod := obj.(*corev1.Pod)
					pod.Name = "bar"
					return admission.Allowed("")
				},
			}, &corev1.Pod{}, decoder)
			result := h.Handle(context.TODO(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: raw,
					},
					Operation: admissionv1.Create,
				},
			})
			Ω(result.Allowed).Should(BeTrue())
			Ω(result.Patches).ShouldNot(BeEmpty())
			result = h.Handle(context.TODO(), admission.Request{})
			Ω(result.Allowed).Should(BeTrue())
		})
		It("should mutate", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
			}
			raw, err := json.Marshal(pod)
			Ω(err).ShouldNot(HaveOccurred())

			h := withMutationHandler(&MutateFunc{
				Func: func(_ context.Context, _ admission.Request, obj runtime.Object) admission.Response {
					pod := obj.(*corev1.Pod)
					pod.Name = "bar"
					return admission.Response{
						AdmissionResponse: admissionv1.AdmissionResponse{
							Allowed: true,
						},
						Patches: []jsonpatch.JsonPatchOperation{},
					}
				},
			}, &corev1.Pod{}, decoder)
			result := h.Handle(context.TODO(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: raw,
					},
					Operation: admissionv1.Create,
				},
			})
			Ω(result.Allowed).Should(BeTrue())
			Ω(result.Patches).Should(BeEmpty())
		})
		It("should validate", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
			}
			raw, err := json.Marshal(pod)
			Ω(err).ShouldNot(HaveOccurred())

			h := withValidationHandler(&ValidateFuncs{
				CreateFunc: func(_ context.Context, _ admission.Request, _ runtime.Object) admission.Response {
					return admission.Allowed("")
				},
				UpdateFunc: func(_ context.Context, _ admission.Request, _ runtime.Object, _ runtime.Object) admission.Response {
					return admission.Denied("")
				},
				DeleteFunc: func(_ context.Context, _ admission.Request, _ runtime.Object) admission.Response {
					return admission.Denied("")
				},
			}, &corev1.Pod{}, decoder)

			result := h.Handle(context.TODO(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: raw,
					},
					Operation: admissionv1.Create,
				},
			})
			Ω(result.Allowed).Should(BeTrue())
			result = h.Handle(context.TODO(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: raw,
					},
					Operation: admissionv1.Update,
				},
			})
			Ω(result.Allowed).Should(BeFalse())
			result = h.Handle(context.TODO(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: raw,
					},
					Operation: admissionv1.Delete,
				},
			})
			Ω(result.Allowed).Should(BeFalse())

		})
		It("should decode object", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
				Spec: corev1.PodSpec{
					NodeName: "jin",
				},
			}
			raw, err := json.Marshal(pod)
			Ω(err).ShouldNot(HaveOccurred())

			h := withMutationHandler(&MutateFunc{
				Func: func(_ context.Context, request admission.Request, object runtime.Object) admission.Response {
					Ω(object).Should(Equal(pod))
					return admission.Allowed("")
				},
			}, &corev1.Pod{}, decoder)

			result := h.Handle(context.TODO(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: raw,
					},
				},
			})
			Ω(result.Allowed).Should(BeTrue())

			result = h.Handle(context.TODO(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: raw,
					},
					OldObject: runtime.RawExtension{
						Raw: raw,
					},
				},
			})
			Ω(result.Allowed).Should(BeTrue())
		})
		It("should not decode invalid object", func() {
			h := withMutationHandler(&MutateFunc{
				Func: func(_ context.Context, _ admission.Request, _ runtime.Object) admission.Response {
					return admission.Allowed("")
				},
			}, &corev1.Pod{}, decoder)

			result := h.Handle(context.TODO(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Object: runtime.RawExtension{
						Raw: []byte{1, 2, 3, 4, 5},
					},
				},
			})
			Ω(result.Allowed).Should(BeFalse())

			result = h.Handle(context.TODO(), admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					OldObject: runtime.RawExtension{
						Raw: []byte{1, 2, 3, 4, 5},
					},
				},
			})
			Ω(result.Allowed).Should(BeFalse())
		})
	})
})
