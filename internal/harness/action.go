package harness

import (
	"context"
	"errors"
	"fmt"

	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/k-harness/operator/internal/executor"
	"github.com/k-harness/operator/internal/harness/checker"
)

type RequestInterface interface {
	Call(ctx context.Context, request executor.Request) (*ActionResult, error)
}

var (
	ErrNoKey       = errors.New("provided key not exists in result")
	ErrBadJsonPath = errors.New("bad json path formula")
)

type Action struct {
	Name string

	v1alpha1.Action
	vars map[string]string
}

func OK() *ActionResult {
	return &ActionResult{Code: "OK"}
}

func NewAction(name string, a v1alpha1.Action, variables map[string]string) *Action {
	return &Action{Name: name, Action: a, vars: variables}
}

// GetBody take stora sync.Map and fill
func (a *Action) GetBody() ([]byte, error) {
	return checker.Body(&a.Request.Body).GetBody(a.vars)
}

func (a *Action) Call(ctx context.Context) (*ActionResult, error) {
	body, err := a.GetBody()
	if err != nil {
		return nil, fmt.Errorf("action can't exstract body: %w", err)
	}

	return Call(ctx, a.Connect, executor.Request{
		Body:   body,
		Type:   a.Request.Body.Type,
		Header: a.Request.Header,
	})
}
