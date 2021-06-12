package checker

import (
	"bytes"
	"fmt"

	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/k-harness/operator/pkg/harness/stuff"
	"github.com/k-harness/operator/pkg/harness/variables"
	"k8s.io/apimachinery/pkg/util/json"
)

type Interface interface {
	Is() error
}

type rs struct {
	vars      *variables.Store
	condition *v1alpha1.ConditionResponse
	response  *stuff.Response
}

func Res(c *v1alpha1.ConditionResponse, v *variables.Store, r *stuff.Response) Interface {
	return &rs{
		vars:      v,
		condition: c,
		response:  r,
	}
}

func (r *rs) Is() error {
	if r.condition.Status != "" {
		if r.condition.Status != r.response.Code {
			return fmt.Errorf("bad status %q expect %q", r.response.Code, r.condition.Status)
		}
	}

	if len(r.condition.Body.KV) > 0 {
		res := make(map[string]interface{})
		if err := json.Unmarshal(r.response.Body, &res); err != nil {
			return fmt.Errorf("can't marshal result for kv check[%w]", err)
		}

		kv := r.vars.TemplateMapOrReturnWhatPossible(r.condition.Body.KV)
		for k, v := range kv {
			rv, ok := res[k]
			if !ok {
				return fmt.Errorf("kv check: key %q not exist", k)
			}

			sv, _ := rv.(string)
			if sv != v {
				return fmt.Errorf("kv check: key %q expect %q got %q", k, v, rv)
			}
		}

		return nil
	}

	body, err := stuff.ScenarioBody(&r.condition.Body).Get()
	if err != nil {
		return fmt.Errorf("get condition Body %w", err)
	}

	if body == nil {
		return nil
	}

	if bytes.Equal(body, r.response.Body) {
		return nil
	}

	return fmt.Errorf("condition %q not equal result %q", string(body), string(r.response.Body))
}
