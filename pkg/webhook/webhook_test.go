package webhook_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/golang/mock/gomock"
	manager "github.com/snorwin/k8s-generic-webhook/pkg/mocks/manager"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	webhook2 "sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/snorwin/k8s-generic-webhook/pkg/webhook"
)

var _ = Describe("Webhook", func() {
	Context("Builder", func() {
		var (
			mock *gomock.Controller
			mgr  *manager.MockManager

			server *webhook2.Server
		)
		BeforeEach(func() {
			mock = gomock.NewController(GinkgoT())
			mgr = manager.NewMockManager(mock)

			scheme := runtime.NewScheme()
			err := corev1.AddToScheme(scheme)
			Ω(err).ShouldNot(HaveOccurred())
			mgr.EXPECT().
				GetScheme().
				Return(scheme).
				AnyTimes()

			mgr.EXPECT().
				GetClient().
				Return(fake.NewClientBuilder().Build()).
				AnyTimes()

			server = &webhook2.Server{}
			mgr.EXPECT().
				GetWebhookServer().
				Return(server).AnyTimes()
		})
		AfterEach(func() {
			mock.Finish()
		})
		It("should build mutating webhook", func() {
			err := webhook.NewGenericWebhookManagedBy(mgr).
				For(&corev1.Pod{}).
				Complete(&webhook.MutatingWebhook{})
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should build validating webhook", func() {
			err := webhook.NewGenericWebhookManagedBy(mgr).
				For(&corev1.Pod{}).
				Complete(&webhook.ValidatingWebhook{})
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should build multiple validating webhook", func() {
			err := webhook.NewGenericWebhookManagedBy(mgr).
				For(&corev1.Pod{}).
				Complete(&webhook.ValidatingWebhook{})
			Ω(err).ShouldNot(HaveOccurred())
			err = webhook.NewGenericWebhookManagedBy(mgr).
				For(&corev1.Namespace{}).
				Complete(&webhook.ValidatingWebhook{})
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should not fail if mutating webhook is already registered", func() {
			err := webhook.NewGenericWebhookManagedBy(mgr).
				For(&corev1.Pod{}).
				Complete(&webhook.MutatingWebhook{})
			Ω(err).ShouldNot(HaveOccurred())

			err = webhook.NewGenericWebhookManagedBy(mgr).
				For(&corev1.Pod{}).
				Complete(&webhook.MutatingWebhook{})
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should not fail if validating webhook is already registered", func() {
			err := webhook.NewGenericWebhookManagedBy(mgr).
				For(&corev1.Pod{}).
				Complete(&webhook.ValidatingWebhook{})
			Ω(err).ShouldNot(HaveOccurred())

			err = webhook.NewGenericWebhookManagedBy(mgr).
				For(&corev1.Pod{}).
				Complete(&webhook.ValidatingWebhook{})
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should not fail if interface doesn't match", func() {
			err := webhook.NewGenericWebhookManagedBy(mgr).
				For(&corev1.Pod{}).
				Complete(struct{}{})
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should fail if api type isn't specified", func() {
			err := webhook.NewGenericWebhookManagedBy(mgr).
				Complete(&webhook.MutatingWebhook{})
			Ω(err).Should(HaveOccurred())
		})
		It("should fail if api type is not registered in scheme", func() {
			err := webhook.NewGenericWebhookManagedBy(mgr).
				For(&appsv1.Deployment{}).
				Complete(&webhook.MutatingWebhook{})
			Ω(err).Should(HaveOccurred())

			err = webhook.NewGenericWebhookManagedBy(mgr).
				For(&appsv1.Deployment{}).
				Complete(&webhook.ValidatingWebhook{})
			Ω(err).Should(HaveOccurred())
		})
	})
})
