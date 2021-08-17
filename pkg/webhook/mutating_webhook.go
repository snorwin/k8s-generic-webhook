package webhook

import (
	"context"
	"encoding/json"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Mutator specifies the interface for a generic mutating webhook.
type Mutator interface {
	// Mutate yields a response to an mutating AdmissionRequest.
	Mutate(context.Context, admission.Request) admission.Response
}

// ensure MutatingWebhook implements Mutator
var _ Mutator = &MutatingWebhook{}

// MutatingWebhook is a generic mutating admission webhook.
type MutatingWebhook struct {
	InjectedClient
	InjectedDecoder
}

// Mutate implements the Mutator interface.
func (m *MutatingWebhook) Mutate(_ context.Context, _ admission.Request) admission.Response {
	return admission.Allowed("")
}

// MutateFunc is a functional interface for a generic mutating admission webhook.
type MutateFunc struct {
	MutatingWebhook

	Func func(context.Context, admission.Request) admission.Response
}

// Mutate implements the Mutator interface by calling the Func.
func (m *MutateFunc) Mutate(ctx context.Context, req admission.Request) admission.Response {
	if m.Func != nil {
		return m.Func(ctx, req)
	}

	return m.MutatingWebhook.Mutate(ctx, req)
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
