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
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Builder builds a Webhook.
type Builder struct {
	mgr            manager.Manager
	apiType        runtime.Object
	pathValidate   string
	pathMutate     string
	prefixValidate string
	prefixMutate   string
}

// NewGenericWebhookManagedBy returns a new webhook Builder that will be invoked by the provided manager.Manager.
func NewGenericWebhookManagedBy(mgr manager.Manager) *Builder {
	return &Builder{
		mgr:            mgr,
		prefixMutate:   "/mutate-",
		prefixValidate: "/validate-",
	}
}

// For takes a runtime.Object which should be a CR.
func (blder *Builder) For(apiType runtime.Object) *Builder {
	blder.apiType = apiType
	return blder
}

// WithMutatePath overrides the mutate path of the webhook
func (blder *Builder) WithMutatePath(path string) *Builder {
	blder.pathMutate = path
	return blder
}

// WithValidatePath overrides the validate path of the webhook
func (blder *Builder) WithValidatePath(path string) *Builder {
	blder.pathValidate = path
	return blder
}

// WithMutatePrefix sets a custom prefix for the mutate path of the webhook, default is '/mutate-'
func (blder *Builder) WithMutatePrefix(prefix string) *Builder {
	blder.prefixMutate = prefix
	return blder
}

// WithValidatePrefix sets a custom prefix for the validate path of the webhook, default is '/validate-'
func (blder *Builder) WithValidatePrefix(prefix string) *Builder {
	blder.prefixMutate = prefix
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

	decoder := admission.NewDecoder(blder.mgr.GetScheme())

	isWebhook := false
	if validator, ok := i.(Validator); ok {
		w := &admission.Webhook{
			Handler: withValidationHandler(validator, blder.apiType, decoder),
		}

		if err := blder.registerValidatingWebhook(w); err != nil {
			return err
		}
		isWebhook = true
	}

	if mutator, ok := i.(Mutator); ok {
		w := &admission.Webhook{
			Handler: withMutationHandler(mutator, blder.apiType, decoder),
		}

		if err := blder.registerMutatingWebhook(w); err != nil {
			return err
		}
		isWebhook = true
	}

	if !isWebhook {
		return fmt.Errorf("webhook instance %v does implement neither Mutator nor Validator interface", i)
	}

	if injector, ok := i.(InjectedClient); ok {
		if err := injector.InjectClient(blder.mgr.GetClient()); err != nil {
			return err
		}
	}

	if injector, ok := i.(InjectedDecoder); ok {
		if err := injector.InjectDecoder(decoder); err != nil {
			return err
		}
	}

	return nil
}

func (blder *Builder) registerValidatingWebhook(w *admission.Webhook) error {
	path := blder.pathValidate
	if strings.TrimSpace(path) == "" {
		gvk, err := apiutil.GVKForObject(blder.apiType, blder.mgr.GetScheme())
		if err != nil {
			return err
		}

		path = generatePath(blder.prefixValidate, gvk)
	}
	if !isAlreadyHandled(blder.mgr, path) {
		blder.mgr.GetWebhookServer().Register(path, w)
	}

	return nil
}

func (blder *Builder) registerMutatingWebhook(w *admission.Webhook) error {
	path := blder.pathMutate
	if strings.TrimSpace(path) == "" {
		gvk, err := apiutil.GVKForObject(blder.apiType, blder.mgr.GetScheme())
		if err != nil {
			return err
		}

		path = generatePath(blder.prefixMutate, gvk)
	}
	if !isAlreadyHandled(blder.mgr, path) {
		blder.mgr.GetWebhookServer().Register(path, w)
	}

	return nil
}

func isAlreadyHandled(mgr ctrl.Manager, path string) bool {
	if mgr.GetWebhookServer().WebhookMux() == nil {
		return false
	}

	h, p := mgr.GetWebhookServer().WebhookMux().Handler(&http.Request{URL: &url.URL{Path: path}})
	if p == path && h != nil {
		return true
	}

	return false
}

func generatePath(prefix string, gvk schema.GroupVersionKind) string {
	return prefix + strings.Replace(gvk.Group, ".", "-", -1) + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}
