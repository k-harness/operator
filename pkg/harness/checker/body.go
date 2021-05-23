package checker

import (
	"fmt"

	"text/template"

	"github.com/google/uuid"
	"github.com/k-harness/operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/rand"
)

var TemplateFunctions = template.FuncMap{
	"uuid": func() string {
		return uuid.New().String()
	},
	"rnd_str": func(len int) string {
		return rand.String(len)
	},
	"range_int": func(min, max int) int {
		return rand.IntnRange(min, max)
	},
}

func Body(in *v1alpha1.Body) *body {
	return &body{Body: in}
}

type body struct {
	*v1alpha1.Body
}

func (b *body) Get() ([]byte, error) {
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