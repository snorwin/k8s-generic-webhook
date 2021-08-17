package webhook

import (
	"context"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"
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

// ensure ValidatingWebhook implements Validator
var _ Validator = &ValidatingWebhook{}

// ValidatingWebhook is a generic validating admission webhook.
type ValidatingWebhook struct {
	InjectedClient
	InjectedDecoder
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

// ValidateObjectFuncs is a functional interface for an object validating admission webhook.
type ValidateObjectFuncs struct {
	ValidatingWebhook

	CreateFunc func(context.Context, admission.Request, runtime.Object) error
	UpdateFunc func(context.Context, admission.Request, runtime.Object, runtime.Object) error
	DeleteFunc func(context.Context, admission.Request, runtime.Object) error
}

// ValidateCreate implements the Validator interface by calling the CreateFunc using the request's runtime.Object.
func (v *ValidateObjectFuncs) ValidateCreate(ctx context.Context, req admission.Request) admission.Response {
	if v.CreateFunc != nil {
		return ValidateCreateObjectByFunc(ctx, req, v.CreateFunc)
	}

	return v.ValidatingWebhook.ValidateCreate(ctx, req)
}

// ValidateUpdate implements the Validator interface by calling the UpdateFunc using the request's runtime.Object.
func (v *ValidateObjectFuncs) ValidateUpdate(ctx context.Context, req admission.Request) admission.Response {
	if v.UpdateFunc != nil {
		return ValidateUpdateObjectByFunc(ctx, req, v.UpdateFunc)
	}

	return v.ValidatingWebhook.ValidateUpdate(ctx, req)
}

// ValidateDelete implements the Validator interface by calling the DeleteFunc using the request's runtime.Object.
func (v *ValidateObjectFuncs) ValidateDelete(ctx context.Context, req admission.Request) admission.Response {
	if v.DeleteFunc != nil {
		return ValidateDeleteObjectByFunc(ctx, req, v.DeleteFunc)
	}

	return v.ValidatingWebhook.ValidateDelete(ctx, req)
}

func ValidateCreateObjectByFunc(ctx context.Context, req admission.Request, f func(context.Context, admission.Request, runtime.Object) error) admission.Response {
	err := f(ctx, req, req.Object.Object)
	if err != nil {
		return admission.Errored(http.StatusForbidden, err)
	}
	return admission.Allowed("")
}

func ValidateUpdateObjectByFunc(ctx context.Context, req admission.Request, f func(context.Context, admission.Request, runtime.Object, runtime.Object) error) admission.Response {
	err := f(ctx, req, req.Object.Object, req.OldObject.Object)
	if err != nil {
		return admission.Errored(http.StatusForbidden, err)
	}
	return admission.Allowed("")
}

func ValidateDeleteObjectByFunc(ctx context.Context, req admission.Request, f func(context.Context, admission.Request, runtime.Object) error) admission.Response {
	err := f(ctx, req, req.Object.Object)
	if err != nil {
		return admission.Errored(http.StatusForbidden, err)
	}
	return admission.Allowed("")
}
