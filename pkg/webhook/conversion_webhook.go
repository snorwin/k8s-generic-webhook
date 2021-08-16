package webhook

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	webhookconversion "sigs.k8s.io/controller-runtime/pkg/webhook/conversion"
)

// Converter specifies the interface for a conversion webhook.
type Converter interface {
	// Convert yields a review to a conversion ConvertionRequest.
	Convert(context.Context, v1beta1.ConversionReview) v1beta1.ConversionResponse
}

// ConversionWebhook is a generic conversion webhook.
type ConversionWebhook struct {
	Client  client.Client
	Decoder *webhookconversion.Decoder
}

// Convert implements the Converter interface.
func (cw *ConversionWebhook) Convert(_ context.Context, _ v1beta1.ConversionRequest) v1beta1.ConversionResponse {
	return v1beta1.ConversionResponse{}
}

// InjectClient implements the inject.Client interface.
func (cw *ConversionWebhook) InjectClient(client client.Client) error {
	cw.Client = client
	return nil
}

// ConvertFunc is a functional interface for a generic conversion conversion webhook.
type ConvertFunc struct {
	ConversionWebhook

	Func func(ctx context.Context, req v1beta1.ConversionRequest) v1beta1.ConversionResponse
}

// Convert implements the Converter interface by calling the Func.
func (cf *ConvertFunc) Convert(ctx context.Context, req v1beta1.ConversionRequest) v1beta1.ConversionResponse {
	if cf.Func != nil {
		return cf.Func(ctx, req)
	}

	return cf.ConversionWebhook.Convert(ctx, req)
}
