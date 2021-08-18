package webhook

import (
	"context"
	"encoding/json"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// ObjectMutator specifies the interface for an object mutating webhook.
type ObjectMutator interface {
	// MutateObject mutates the runtime.Object of a admission.Request
	MutateObject(context.Context, admission.Request, runtime.Object) error
}

// wrapAsMutator wrap the given ObjectMutator as generic Mutator
func wrapAsMutator(mutator ObjectMutator) Mutator {
	return &objectMutatorAdapter{mutator}
}

type objectMutatorAdapter struct {
	ObjectMutator
}

// Mutate implements the Mutator interface.
func (a *objectMutatorAdapter) Mutate(ctx context.Context, req admission.Request) admission.Response {
	return MutateObjectByFunc(ctx, req, a.ObjectMutator.MutateObject)
}

// InjectDecoder implements the admission.DecoderInjector interface.
func (a *objectMutatorAdapter) InjectDecoder(decoder *admission.Decoder) error {
	// pass decoder to the underlying handler
	if injector, ok := a.ObjectMutator.(admission.DecoderInjector); ok {
		return injector.InjectDecoder(decoder)
	}

	return nil
}

// InjectClient implements the inject.Client interface.
func (a *objectMutatorAdapter) InjectClient(client client.Client) error {
	// pass client to the underlying handler
	if injector, ok := a.ObjectMutator.(inject.Client); ok {
		return injector.InjectClient(client)
	}

	return nil
}

// ensure MutatingObjectWebhook implements ObjectMutator
var _ ObjectMutator = &MutatingObjectWebhook{}

// MutatingObjectWebhook is a simplified mutating admission webhook.
type MutatingObjectWebhook struct {
	InjectedClient
	InjectedDecoder
}

// MutateObject implements the ObjectMutator interface.
func (w *MutatingObjectWebhook) MutateObject(_ context.Context, _ admission.Request, _ runtime.Object) error {
	return nil
}

// MutateObjectFunc is a functional interface for an object mutating admission webhook.
type MutateObjectFunc struct {
	MutatingWebhook

	Func func(context.Context, admission.Request, runtime.Object) error
}

// Mutate implements the Mutator interface by calling the Func using the request's runtime.Object.
func (m *MutateObjectFunc) Mutate(ctx context.Context, req admission.Request) admission.Response {
	if m.Func != nil {
		return MutateObjectByFunc(ctx, req, m.Func)
	}

	return m.MutatingWebhook.Mutate(ctx, req)
}

// MutateObjectByFunc yields and admission.Response for a runtime.Object mutated bu the specified function.
func MutateObjectByFunc(ctx context.Context, req admission.Request, f func(context.Context, admission.Request, runtime.Object) error) admission.Response {
	obj := req.Object.Object
	err := f(ctx, req, obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	marshalled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshalled)
}
