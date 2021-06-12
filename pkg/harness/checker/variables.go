package checker

import (
	"fmt"

	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/k-harness/operator/pkg/harness/variables"
)

type vars struct {
	store     *variables.Store
	condition *v1alpha1.ConditionVariables
}

func Vars(c *v1alpha1.ConditionVariables, v *variables.Store) Interface {
	return &vars{store: v, condition: c}
}

func (v *vars) Is() error {
	for key, check := range v.condition.KV {
		v, ok := v.store.Get(key)
		if !ok {
			return fmt.Errorf("key %q not exists in store", key)
		}

		switch check.Operator {
		case v1alpha1.Required:
		case v1alpha1.Equal:
			if check.Value != v {
				return fmt.Errorf("key %q  store value %q not equal desired value %q",
					key, v, check.Value)
			}
		}
	}

	return nil
}
