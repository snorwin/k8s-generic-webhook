package webhook

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// SimpleValidator specifies the interface for a simple validating webhook.
type SimpleValidator interface {
	// ValidateCreate yields a response to an validating AdmissionRequest with operation set to Create.
	ValidateCreate(context.Context, admission.Request, runtime.Object) error
	// ValidateUpdate yields a response to an validating AdmissionRequest with operation set to Update.
	ValidateUpdate(context.Context, admission.Request, runtime.Object, runtime.Object) error
	// ValidateDelete yields a response to an validating AdmissionRequest with operation set to Delete.
	ValidateDelete(context.Context, admission.Request, runtime.Object) error
}

// wrapAsValidator wrap the given simple validator as generic validator
func wrapAsValidator(sv SimpleValidator) Validator {
	return &simpleValidatorAdapter{
		ValidatingWebhook: ValidatingWebhook{
			baseHandler{child: sv},
		},
	}
}

type simpleValidatorAdapter struct {
	ValidatingWebhook
	sm SimpleValidator
}

func (s *simpleValidatorAdapter) ValidateCreate(ctx context.Context, request admission.Request) admission.Response {
	err := s.sm.ValidateCreate(ctx, request, request.Object.Object)
	if err != nil {
		return admission.Errored(http.StatusForbidden, err)
	}
	return admission.Allowed("")
}

func (s *simpleValidatorAdapter) ValidateUpdate(ctx context.Context, request admission.Request) admission.Response {
	err := s.sm.ValidateUpdate(ctx, request, request.Object.Object, request.OldObject.Object)
	if err != nil {
		return admission.Errored(http.StatusForbidden, err)
	}
	return admission.Allowed("")
}

func (s *simpleValidatorAdapter) ValidateDelete(ctx context.Context, request admission.Request) admission.Response {
	err := s.sm.ValidateDelete(ctx, request, request.Object.Object)
	if err != nil {
		return admission.Errored(http.StatusForbidden, err)
	}
	return admission.Allowed("")
}

// ensure SimpleValidatingWebhook implements SimpleValidator
var _ SimpleValidator = &SimpleValidatingWebhook{}

// SimpleValidatingWebhook is a generic validating admission webhook.
type SimpleValidatingWebhook struct {
	baseHandler
}

// ValidateCreate implements the SimpleValidator interface.
func (v *SimpleValidatingWebhook) ValidateCreate(_ context.Context, _ admission.Request, _ runtime.Object) error {
	return nil
}

// ValidateUpdate implements the SimpleValidator interface.
func (v *SimpleValidatingWebhook) ValidateUpdate(_ context.Context, _ admission.Request, _ runtime.Object, _ runtime.Object) error {
	return nil
}

// ValidateDelete implements the SimpleValidator interface.
func (v *SimpleValidatingWebhook) ValidateDelete(_ context.Context, _ admission.Request, _ runtime.Object) error {
	return nil
}

// SimpleValidateFuncs is a functional interface for a simple validating admission webhook.
type SimpleValidateFuncs struct {
	SimpleValidatingWebhook

	CreateFunc func(context.Context, admission.Request, runtime.Object) error
	UpdateFunc func(context.Context, admission.Request, runtime.Object, runtime.Object) error
	DeleteFunc func(context.Context, admission.Request, runtime.Object) error
}

// ValidateCreate implements the Validator interface by calling the CreateFunc.
func (v *SimpleValidateFuncs) ValidateCreate(ctx context.Context, req admission.Request, obj runtime.Object) error {
	if v.CreateFunc != nil {
		return v.CreateFunc(ctx, req, obj)
	}

	return v.SimpleValidatingWebhook.ValidateCreate(ctx, req, obj)
}

// ValidateUpdate implements the SimpleValidator interface by calling the UpdateFunc.
func (v *SimpleValidateFuncs) ValidateUpdate(ctx context.Context, req admission.Request, objNew runtime.Object, objOld runtime.Object) error {
	if v.UpdateFunc != nil {
		return v.UpdateFunc(ctx, req, objNew, objOld)
	}

	return v.SimpleValidatingWebhook.ValidateUpdate(ctx, req, objNew, objOld)
}

// ValidateDelete implements the SimpleValidator interface by calling the DeleteFunc.
func (v *SimpleValidateFuncs) ValidateDelete(ctx context.Context, req admission.Request, obj runtime.Object) error {
	if v.DeleteFunc != nil {
		return v.DeleteFunc(ctx, req, obj)
	}

	return v.SimpleValidatingWebhook.ValidateDelete(ctx, req, obj)
}
