package webhook

import (
	"fmt"
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
	"sigs.k8s.io/controller-runtime/pkg/webhook/conversion"
)

// Builder builds a Webhook.
type Builder struct {
	mgr            manager.Manager
	apiType        runtime.Object
	pathValidate   string
	pathMutate     string
	pathConvert    string
	prefixValidate string
	prefixMutate   string
	prefixConvert  string
}

// NewGenericWebhookManagedBy returns a new webhook Builder that will be invoked by the provided manager.Manager.
func NewGenericWebhookManagedBy(mgr manager.Manager) *Builder {
	return &Builder{
		mgr:            mgr,
		prefixMutate:   "/mutate-",
		prefixValidate: "/validate-",
		prefixConvert:  "/convert-",
	}
}

// For takes a runtime.Object which should be a CR.
func (blder *Builder) For(apiType runtime.Object) *Builder {
	blder.apiType = apiType
	return blder
}

func (blder *Builder) WithMutatePath(path string) *Builder {
	blder.pathMutate = path
	return blder
}

func (blder *Builder) WithValidatePath(path string) *Builder {
	blder.pathValidate = path
	return blder
}

func (blder *Builder) WithConvertPath(path string) *Builder {
	blder.pathConvert = path
	return blder
}

func (blder *Builder) WithMutatePrefix(prefix string) *Builder {
	blder.prefixMutate = prefix
	return blder
}

func (blder *Builder) WithValidatePrefix(prefix string) *Builder {
	blder.prefixMutate = prefix
	return blder
}

func (blder *Builder) WithConvertPrefix(prefix string) *Builder {
	blder.prefixConvert = prefix
	return blder
}

// Complete builds the webhook.
// If the given object implements the Mutator interface, a MutatingWebhook will be created.
// If the given object implements the Validator interface, a ValidatingWebhook will be created.
func (blder *Builder) Complete(i interface{}) error {

	if blder.pathMutate != "" && !strings.HasPrefix(blder.pathMutate, "/") {
		return fmt.Errorf("mutating path %q must start with '/'", blder.pathMutate)
	} else if !strings.HasPrefix(blder.prefixMutate, "/") {
		return fmt.Errorf("mutating prefix %q must start with '/'", blder.prefixMutate)
	}
	if blder.pathValidate != "" && !strings.HasPrefix(blder.pathValidate, "/") {
		return fmt.Errorf("validating path %q must start with '/'", blder.pathValidate)
	} else if !strings.HasPrefix(blder.prefixValidate, "/") {
		return fmt.Errorf("validating prefix %q must start with '/'", blder.prefixValidate)
	}
	if blder.pathConvert != "" && !strings.HasPrefix(blder.pathConvert, "/") {
		return fmt.Errorf("conversion path %q must start with '/'", blder.pathConvert)
	} else if !strings.HasPrefix(blder.prefixConvert, "/") {
		return fmt.Errorf("conversion prefix %q must start with '/'", blder.prefixConvert)
	}

	if validator, ok := i.(Validator); ok {
		w, err := blder.createAdmissionWebhook(&handler{Handler: validator, Object: blder.apiType})
		if err != nil {
			return err
		}

		if err := blder.registerValidatingWebhook(w); err != nil {
			return err
		}
	}

	if mutator, ok := i.(Mutator); ok {
		w, err := blder.createAdmissionWebhook(&handler{Handler: mutator, Object: blder.apiType})
		if err != nil {
			return err
		}

		if err := blder.registerMutatingWebhook(w); err != nil {
			return err
		}
	}

	if converter, ok := i.(Converter); ok {
		w, err := blder.createConversionWebhook(&handler{Handler: converter, Object: blder.apiType})
		if err != nil {
			return err
		}

		if err := blder.registerConversionWebhook(w); err != nil {
			return err
		}
	}

	return nil
}

func (blder *Builder) createAdmissionWebhook(handler Handler) (*admission.Webhook, error) {
	w := &admission.Webhook{
		Handler:         handler,
		WithContextFunc: nil,
	}

	// inject scheme for decoder
	if err := w.InjectScheme(blder.mgr.GetScheme()); err != nil {
		return nil, err
	}

	// inject client
	if err := w.InjectFunc(func(i interface{}) error {
		if injector, ok := i.(inject.Client); ok {
			return injector.InjectClient(blder.mgr.GetClient())
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

	path := generatePath(blder.pathValidate, blder.prefixValidate, gvk)
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

	path := generatePath(blder.pathMutate, blder.prefixMutate, gvk)
	if !isAlreadyHandled(blder.mgr, path) {
		blder.mgr.GetWebhookServer().Register(path, w)
	}

	return nil
}

func (blder *Builder) createConversionWebhook(handler Handler) (*conversion.Webhook, error) {
	w := &conversion.Webhook{}

	// inject scheme for decoder
	if err := w.InjectScheme(blder.mgr.GetScheme()); err != nil {
		return nil, err
	}

	return w, nil
}

func (blder *Builder) registerConversionWebhook(w *conversion.Webhook) error {
	gvk, err := apiutil.GVKForObject(blder.apiType, blder.mgr.GetScheme())
	if err != nil {
		return err
	}

	path := generatePath(blder.pathConvert, blder.prefixConvert, gvk)
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

func generatePath(override string, prefix string, gvk schema.GroupVersionKind) string {
	if override != "" {
		return override
	}
	return prefix + strings.Replace(gvk.Group, ".", "-", -1) + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}
