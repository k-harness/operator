package models

import (
	"github.com/k-harness/operator/api/v1alpha1"
)

type Action struct {
	Name string

	v1alpha1.Action
}

func NewAction(name string, a v1alpha1.Action) *Action {
	return &Action{Name: name, Action: a}
}

// GetBody take stora sync.Map and fill
func (a *Action) GetBody(store map[string]string) ([]byte, error) {
	return Body(&a.Request.Body).GetBody(store)
}
