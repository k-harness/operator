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
	if r.condition.Status != nil && *r.condition.Status != r.response.Code {
		return fmt.Errorf("bad status %q expect %q", r.response.Code, *r.condition.Status)
	}

	for key, check := range r.condition.JSONPath {
		v, err := r.response.GetKeyValue(key)
		if err != nil {
			return fmt.Errorf("ConditionResponse:JSONPath %q err: %w", check.Value, err)
		}

		if check.Operator == v1alpha1.Equal {
			if check.Value != v {
				return fmt.Errorf("ConditionResponse:JSONPath %q => %q != %q", check.Value, v, check.Value)
			}
		}
	}

	return r.bodyCheck()
}

func (r *rs) bodyCheck() error {
	if r.condition.Body == nil {
		return nil
	}

	body := r.condition.Body

	if len(body.KV) > 0 {
		res := make(map[string]interface{})
		if err := json.Unmarshal(r.response.Body, &res); err != nil {
			return fmt.Errorf("can't marshal result for kv check[%w]", err)
		}

		kv := r.vars.TemplateMapOrReturnWhatPossible(body.KV)
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

	bb, err := stuff.ScenarioBody(body).Get()
	if err != nil {
		return fmt.Errorf("get condition Body %w", err)
	}

	if bb == nil {
		return nil
	}

	if bytes.Equal(bb, r.response.Body) {
		return nil
	}

	return fmt.Errorf("condition %q not equal result %q", string(bb), string(r.response.Body))
}
