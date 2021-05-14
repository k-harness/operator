package checker

import (
	"bytes"
	"fmt"

	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/k-harness/operator/internal/harness/models"
)

type ResCheckInterface interface {
	Is(status string, res []byte) error
}

type rs struct {
	store     map[string]string
	condition *v1alpha1.ConditionResponse
}

func ResCheck(store map[string]string, condition *v1alpha1.ConditionResponse) ResCheckInterface {
	return &rs{
		store:     store,
		condition: condition,
	}
}

func (r *rs) Is(status string, resBody []byte) error {
	if r.condition.Status != "" {
		if r.condition.Status != status {
			return fmt.Errorf("bad status %q expect %q", status, r.condition.Status)
		}
	}

	body, err := models.Body(&r.condition.Body).GetBody(r.store)
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
