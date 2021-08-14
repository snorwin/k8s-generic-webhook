package webhook

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Validator specifies the interface for a validating webhook.
type Validator interface {
	// ValidateCreate yields a response to an validating AdmissionRequest with operation set to Create.
	ValidateCreate(context.Context, admission.Request) admission.Response
	// ValidateUpdate yields a response to an validating AdmissionRequest with operation set to Update.
	ValidateUpdate(context.Context, admission.Request) admission.Response
	// ValidateDelete yields a response to an validating AdmissionRequest with operation set to Delete.
	ValidateDelete(context.Context, admission.Request) admission.Response
}

// ValidatingWebhook is a generic validating admission webhook.
type ValidatingWebhook struct {
	Client  client.Client
	Decoder *admission.Decoder
}

// ValidateCreate implements the Validator interface.
func (v *ValidatingWebhook) ValidateCreate(_ context.Context, _ admission.Request) admission.Response {
	return admission.Allowed("")
}

// ValidateUpdate implements the Validator interface.
func (v *ValidatingWebhook) ValidateUpdate(_ context.Context, _ admission.Request) admission.Response {
	return admission.Allowed("")
}

// ValidateDelete implements the Validator interface.
func (v *ValidatingWebhook) ValidateDelete(_ context.Context, _ admission.Request) admission.Response {
	return admission.Allowed("")
}

// InjectDecoder implements the admission.DecoderInjector interface.
func (v *ValidatingWebhook) InjectDecoder(decoder *admission.Decoder) error {
	v.Decoder = decoder
	return nil
}

// InjectClient implements the inject.Client interface.
func (v *ValidatingWebhook) InjectClient(client client.Client) error {
	v.Client = client
	return nil
}

// ValidateFuncs is a functional interface for a generic validating admission webhook.
type ValidateFuncs struct {
	ValidatingWebhook

	CreateFunc func(context.Context, admission.Request) admission.Response
	UpdateFunc func(context.Context, admission.Request) admission.Response
	DeleteFunc func(context.Context, admission.Request) admission.Response
}

// ValidateCreate implements the Validator interface by calling the CreateFunc.
func (v *ValidateFuncs) ValidateCreate(ctx context.Context, req admission.Request) admission.Response {
	if v.CreateFunc != nil {
		return v.CreateFunc(ctx, req)
	}

	return v.ValidatingWebhook.ValidateCreate(ctx, req)
}

// ValidateUpdate implements the Validator interface by calling the UpdateFunc.
func (v *ValidateFuncs) ValidateUpdate(ctx context.Context, req admission.Request) admission.Response {
	if v.UpdateFunc != nil {
		return v.UpdateFunc(ctx, req)
	}

	return v.ValidatingWebhook.ValidateUpdate(ctx, req)
}

// ValidateDelete implements the Validator interface by calling the DeleteFunc.
func (v *ValidateFuncs) ValidateDelete(ctx context.Context, req admission.Request) admission.Response {
	if v.DeleteFunc != nil {
		return v.DeleteFunc(ctx, req)
	}

	return v.ValidatingWebhook.ValidateDelete(ctx, req)
}
