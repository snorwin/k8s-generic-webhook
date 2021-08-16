package webhook

import (
	"context"
	"encoding/json"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SimpleMutator specifies the interface for a simple mutating webhook where the current object is presentes as agument and patch generation is automated.
type SimpleMutator interface {
	// Mutate yields a response to an mutating AdmissionRequest with the given target object.
	Mutate(context.Context, admission.Request, runtime.Object) error
}

// wrapAsMutator wrap the given simple mutator as generic mutator
func wrapAsMutator(sm SimpleMutator) Mutator {
	return &simpleMutatorAdapter{
		MutatingWebhook: MutatingWebhook{
			baseHandler{child: sm},
		},
	}
}

type simpleMutatorAdapter struct {
	MutatingWebhook
	sm SimpleMutator
}

func (s *simpleMutatorAdapter) Mutate(ctx context.Context, req admission.Request) admission.Response {
	// invoke object mutator
	obj := req.Object.Object
	err := s.sm.Mutate(ctx, req, obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	marshalled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshalled)
}

// ensure SimpleMutatingWebhook implements SimpleMutator
var _ SimpleMutator = &SimpleMutatingWebhook{}

// SimpleMutatingWebhook is a simple mutating admission webhook.
type SimpleMutatingWebhook struct {
	baseHandler
}

// Mutate implements the SimpleMutator interface.
func (s *SimpleMutatingWebhook) Mutate(_ context.Context, _ admission.Request, _ runtime.Object) error {
	return nil
}

// SimpleMutateFunc is a functional interface for a simple mutating admission webhook.
type SimpleMutateFunc struct {
	SimpleMutatingWebhook
	Func func(context.Context, admission.Request, runtime.Object) error
}

// Mutate implements the SimpleMutator interface by calling the Func.
func (m *SimpleMutateFunc) Mutate(ctx context.Context, req admission.Request, obj runtime.Object) error {
	if m.Func != nil {
		return m.Func(ctx, req, obj)
	}

	return m.SimpleMutatingWebhook.Mutate(ctx, req, obj)
}
