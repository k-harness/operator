package stuff

import (
	"fmt"

	"github.com/k-harness/operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/json"
)

func ScenarioBody(in *v1alpha1.Body) *Body {
	return &Body{Body: in}
}

type Body struct {
	*v1alpha1.Body
}

func (b *Body) Get() ([]byte, error) {
	if len(b.Byte) > 0 {
		return b.Byte, nil
	}

	if len(b.Row) > 0 {
		return []byte(b.Row), nil
	}

	if len(b.KV) == 0 {
		return nil, nil
	}

	body, err := json.Marshal(b.KV)
	if err != nil {
		return nil, fmt.Errorf("kv marshal error: %w", err)
	}

	return body, nil
}
