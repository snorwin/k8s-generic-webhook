package webhook

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// ClientInjector is used to inject a client.Client into webhook handlers.
type ClientInjector interface {
	InjectClient(client.Client) error
}

// ensure InjectedClient implements ClientInjector
var _ ClientInjector = &InjectedClient{}

// InjectedClient holds an injected client.Client
type InjectedClient struct {
	Client client.Client
}

// InjectClient implements the ClientInjector interface.
func (i *InjectedClient) InjectClient(client client.Client) error {
	i.Client = client
	return nil
}

// DecoderInjector is used to inject an admission.Decoder into webhook handlers.
type DecoderInjector interface {
	InjectDecoder(admission.Decoder) error
}

// ensure InjectedDecoder implements DecoderInjector
var _ DecoderInjector = &InjectedDecoder{}

// InjectedDecoder holds an injected admission.Decoder
type InjectedDecoder struct {
	Decoder admission.Decoder
}

// InjectDecoder implements the DecoderInjector interface.
func (i *InjectedDecoder) InjectDecoder(decoder admission.Decoder) error {
	i.Decoder = decoder
	return nil
}
