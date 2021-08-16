package webhook_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/snorwin/k8s-generic-webhook/pkg/webhook"
)

// TODO: properly test if conversion base interfaces are satisfied after initial merge

// isHub determines if passed-in object is a Hub or not.
// func isHub(obj runtime.Object) bool {
// 	_, yes := obj.(conversion.Hub)
// 	return yes
// }

// // isConvertible determines if passed-in object is a convertible.
// func isConvertible(obj runtime.Object) bool {
// 	_, yes := obj.(conversion.Convertible)
// 	return yes
// }

var _ = Describe("Conversion webhook"), func() {
	Context("ConvertFunc", func() {
		It("should be sucessfull", func() {
			result := (&webhook.ConvertFunc{}).Convert(context.TODO(), v1beta1.ConversionRequest{})
			Ω(result.Result.Status).Should(Equal("Success"))
		})
		// It("should be Convertible", func() {
		// TODO
		// })
		// It("should be Hub ", func() {
		// TODO
		// })
	})
}
