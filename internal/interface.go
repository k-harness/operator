package internal

import (
	"context"

	"github.com/k-harness/operator/api/v1alpha1"
)

type Kube interface {
	Update(item *v1alpha1.Scenario) error
}

type HarnessFactory interface {
	// Add should be called via Kube controller when object created
	Add(ctx context.Context, add interface{})
}
