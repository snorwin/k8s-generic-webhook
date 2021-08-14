package webhook

import (
	"net/http"
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Builder builds a Webhook.
type Builder struct {
	mgr     manager.Manager
	apiType runtime.Object
}

// NewGenericWebhookManagedBy returns a new webhook Builder that will be invoked by the provided manager.Manager.
func NewGenericWebhookManagedBy(mgr manager.Manager) *Builder {
	return &Builder{mgr: mgr}
}

// For takes a runtime.Object which should be a CR.
func (bldr *Builder) For(apiType runtime.Object) *Builder {
	bldr.apiType = apiType
	return bldr
}

// Complete builds the webhook.
// If the given object implements the Mutator interface, a MutatingWebhook will be created.
// If the given object implements the Validator interface, a ValidatingWebhook will be created.
func (bldr *Builder) Complete(i interface{}) error {
	if validator, ok := i.(Validator); ok {
		w, err := bldr.createAdmissionWebhook(&handler{Handler: validator, Object: bldr.apiType})
		if err != nil {
			return err
		}

		if err := bldr.registerValidatingWebhook(w); err != nil {
			return err
		}
	}

	if mutator, ok := i.(Mutator); ok {
		w, err := bldr.createAdmissionWebhook(&handler{Handler: mutator, Object: bldr.apiType})
		if err != nil {
			return err
		}

		if err := bldr.registerValidatingWebhook(w); err != nil {
			return err
		}
	}

	return nil
}

func (bldr *Builder) createAdmissionWebhook(handler Handler) (*admission.Webhook, error) {
	w := &admission.Webhook{
		Handler:         handler,
		WithContextFunc: nil,
	}

	// inject scheme for decoder
	if err := w.InjectScheme(bldr.mgr.GetScheme()); err != nil {
		return nil, err
	}

	// inject client
	if err := w.InjectFunc(func(i interface{}) error {
		if injector, ok := i.(inject.Client); ok {
			return injector.InjectClient(bldr.mgr.GetClient())
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return w, nil
}
func (blder *Builder) registerValidatingWebhook(w *admission.Webhook) error {
	gvk, err := apiutil.GVKForObject(blder.apiType, blder.mgr.GetScheme())
	if err != nil {
		return err
	}

	path := generatePath("/validate-", gvk)
	if !isAlreadyHandled(blder.mgr, path) {
		blder.mgr.GetWebhookServer().Register(path, w)
	}

	return nil
}

func (blder *Builder) registerMutatingWebhook(w *admission.Webhook) error {
	gvk, err := apiutil.GVKForObject(blder.apiType, blder.mgr.GetScheme())
	if err != nil {
		return err
	}

	path := generatePath("/mutate-", gvk)
	if !isAlreadyHandled(blder.mgr, path) {
		blder.mgr.GetWebhookServer().Register(path, w)
	}

	return nil
}

func isAlreadyHandled(mgr ctrl.Manager, path string) bool {
	if mgr.GetWebhookServer().WebhookMux == nil {
		return false
	}

	h, p := mgr.GetWebhookServer().WebhookMux.Handler(&http.Request{URL: &url.URL{Path: path}})
	if p == path && h != nil {
		return true
	}

	return false
}

func generatePath(prefix string, gvk schema.GroupVersionKind) string {
	return prefix + strings.Replace(gvk.Group, ".", "-", -1) + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}
