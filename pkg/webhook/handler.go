package webhook

import (
	"context"
	"encoding/json"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// withValidationHandler create a validation handler instance
func withValidationHandler(validator Validator, object runtime.Object, decoder admission.Decoder) admission.Handler {
	return &handler{validator: validator, Object: object, decoder: decoder}
}

// withMutationHandler create a mutation handler instance
func withMutationHandler(mutator Mutator, object runtime.Object, decoder admission.Decoder) admission.Handler {
	return &handler{mutator: mutator, Object: object, decoder: decoder}
}

// handler is wrapper type for Validator and Mutator, implements the Handler interface.
type handler struct {
	// validator instance, should be nil if mutator is set
	validator Validator
	// mutator instance, should be nil if validator is set
	mutator Mutator

	Object runtime.Object

	decoder admission.Decoder
}

// Handle implements the admission.Handler interface.
func (h *handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	// add metadata to context's logger
	logger := log.FromContext(ctx).
		WithValues("name", req.Name).
		WithValues("namespace", req.Namespace).
		WithValues("gvk", req.Kind.String()).
		WithValues("uid", req.UID)
	ctx = log.IntoContext(ctx, logger)

	// decode object
	if len(req.Object.Raw) > 0 && req.Object.Object == nil {
		obj := h.Object.DeepCopyObject()
		if err := h.decoder.DecodeRaw(req.Object, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		req.Object.Object = obj
	}

	// decode old object
	if len(req.OldObject.Raw) > 0 && req.OldObject.Object == nil {
		obj := h.Object.DeepCopyObject()
		if err := h.decoder.DecodeRaw(req.OldObject, obj); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		req.OldObject.Object = obj
	}

	// invoke validator
	if h.validator != nil {
		switch req.Operation {
		case admissionv1.Create:
			return h.validator.ValidateCreate(ctx, req, req.Object.Object)
		case admissionv1.Update:
			return h.validator.ValidateUpdate(ctx, req, req.Object.Object, req.OldObject.Object)
		case admissionv1.Delete:
			return h.validator.ValidateDelete(ctx, req, req.OldObject.Object)
		}
	}

	// invoke mutator
	if h.mutator != nil {
		if req.Object.Object != nil {
			resp := h.mutator.Mutate(ctx, req, req.Object.Object)
			if resp.Allowed && resp.Patches == nil {
				// generate patches
				marshalled, err := json.Marshal(req.Object.Object)
				if err != nil {
					return admission.Errored(http.StatusInternalServerError, err)
				}

				return admission.PatchResponseFromRaw(req.Object.Raw, marshalled)
			}

			return resp
		} else {
			return admission.Allowed("")
		}
	}

	return admission.Denied("")
}
