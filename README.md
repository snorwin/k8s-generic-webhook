# k8s-generic-webhook

[![GitHub Action](https://img.shields.io/badge/GitHub-Action-blue)](https://github.com/features/actions)
[![Documentation](https://img.shields.io/badge/godoc-reference-5272B4.svg)](https://pkg.go.dev/github.com/snorwin/k8s-generic-webhook/pkg/webhook)
[![Test](https://img.shields.io/github/workflow/status/snorwin/k8s-generic-webhook/Test?label=tests&logo=github)](https://github.com/snorwin/k8s-generic-webhook/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/snorwin/k8s-generic-webhook)](https://goreportcard.com/report/github.com/snorwin/k8s-generic-webhook)
[![Coverage Status](https://coveralls.io/repos/github/snorwin/k8s-generic-webhook/badge.svg?branch=main)](https://coveralls.io/github/snorwin/k8s-generic-webhook?branch=main)
[![Releases](https://img.shields.io/github/v/release/snorwin/k8s-generic-webhook)](https://github.com/snorwin/k8s-generic-webhook/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

The **k8s-generic-webhook** is a library to simplify the implementation of webhooks for arbitrary customer resources (CR) in the [operator-sdk](https://sdk.operatorframework.io/) or [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime).
Furthermore, it provides full access to the `AdmissionReview` request and decodes the `Object` in the request automatically. More sophistic webhook logic is facilitated by using the injected `Client` of the webhook which provides full access to the Kubernetes API.

## Quickstart
1. Initialize a new manager using the [operator-sdk](https://sdk.operatorframework.io/).
2. Create a pkg (e.g. `webhooks/pod`) and implement your webhook logic by embedding either the `ValidatingWebhook` or the `MuatatingWebhook`.

#### Example `ValidatingWebhook`
```go
package pod

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/snorwin/k8s-generic-webhook/pkg/webhook"
)

type Webhook struct {
	webhook.ValidatingWebhook
}

func (w *Webhook) SetupWebhookWithManager(mgr manager.Manager) error {
	return webhook.NewGenericWebhookManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(w)
}

func (w *Webhook) ValidateCreate(ctx context.Context, request admission.Request, object runtime.Object) admission.Response {
	_ = log.FromContext(ctx)

	pod := object.(*corev1.Pod)
	// TODO add your programmatic validation logic here

	return admission.Allowed("")
}
```

#### Example `MutatingWebhook`
```go
package pod

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/snorwin/k8s-generic-webhook/pkg/webhook"
)

type Webhook struct {
	webhook.MutatingWebhook
}

func (w *Webhook) SetupWebhookWithManager(mgr manager.Manager) error {
	return webhook.NewGenericWebhookManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(&w)
}

func (w *Webhook) Mutate(ctx context.Context, request admission.Request, object runtime.Object) admission.Response {
	_ = log.FromContext(ctx)

	pod := object.(*corev1.Pod)
	// TODO add your programmatic mutation logic here

	return admission.Allowed("")
}
```

3. Add the following snippet to `main()` in `main.go` in order to register the webhook in the manager.
```go
if err = (&pod.Webhook{}).SetupWebhookWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create webhook", "webhook", "Pod")
    os.Exit(1)
}
```
