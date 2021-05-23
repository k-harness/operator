package checker

import (
	"bytes"
	"fmt"

	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/k-harness/operator/pkg/harness/variables"
	"k8s.io/apimachinery/pkg/util/json"
)

type ResCheckInterface interface {
	Is(status string, res []byte) error
}

type rs struct {
	vars      *variables.Store
	condition *v1alpha1.ConditionResponse
}

func ResCheck(vars *variables.Store, condition *v1alpha1.ConditionResponse) ResCheckInterface {
	return &rs{
		vars:      vars,
		condition: condition,
	}
}

func (r *rs) Is(status string, resBody []byte) error {
	if r.condition.Status != "" {
		if r.condition.Status != status {
			return fmt.Errorf("bad status %q expect %q", status, r.condition.Status)
		}
	}

	if len(r.condition.Body.KV) > 0 {
		res := make(map[string]interface{})
		if err := json.Unmarshal(resBody, &res); err != nil {
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

	body, err := Body(&r.condition.Body).Get()
	if err != nil {
		return fmt.Errorf("get condition body %w", err)
	}

	if body == nil {
		return nil
	}

	if bytes.Equal(body, resBody) {
		return nil
	}

	return fmt.Errorf("condition %q not equal result %q", string(body), string(resBody))
}
