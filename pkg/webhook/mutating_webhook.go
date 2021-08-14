package webhook

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Mutator specifies the interface for a mutating webhook.
type Mutator interface {
	// Mutate yields a response to an mutating AdmissionRequest.
	Mutate(context.Context, admission.Request) admission.Response
}

// MutatingWebhook is a generic mutating admission webhook.
type MutatingWebhook struct {
	Client  client.Client
	Decoder *admission.Decoder
}

// Mutate implements the Mutator interface.
func (m *MutatingWebhook) Mutate(_ context.Context, _ admission.Request) admission.Response {
	return admission.Allowed("")
}

// InjectDecoder implements the admission.DecoderInjector interface.
func (m *MutatingWebhook) InjectDecoder(decoder *admission.Decoder) error {
	m.Decoder = decoder
	return nil
}

// InjectClient implements the inject.Client interface.
func (m *MutatingWebhook) InjectClient(client client.Client) error {
	m.Client = client
	return nil
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
