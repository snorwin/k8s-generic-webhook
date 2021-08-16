package webhook

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// baseHandler is a generic mutating admission webhook.
type baseHandler struct {
	Client  client.Client
	Decoder *admission.Decoder

	child interface{}
}

// InjectDecoder implements the admission.DecoderInjector interface.
func (b *baseHandler) InjectDecoder(decoder *admission.Decoder) error {
	b.Decoder = decoder
	// pass decoder to the underlying handler
	if injector, ok := b.child.(admission.DecoderInjector); ok {
		return injector.InjectDecoder(decoder)
	}
	return nil
}

// InjectClient implements the inject.Client interface.
func (b *baseHandler) InjectClient(client client.Client) error {
	b.Client = client
	// pass client to the underlying handler
	if injector, ok := b.child.(inject.Client); ok {
		return injector.InjectClient(client)
	}
	return nil
}
