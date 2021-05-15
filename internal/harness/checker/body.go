package checker

import (
	"bytes"
	"fmt"

	"text/template"

	"github.com/google/uuid"
	"github.com/k-harness/operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/json"
)

var TemplateFunctions = template.FuncMap{
	"uuid": func() string {
		return uuid.New().String()
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

func (b *body) GetBody(store map[string]string) ([]byte, error) {
	res, err := b.Get()
	if err != nil {
		return nil, err
	}

	t, err := template.New("x").
		Funcs(TemplateFunctions).
		Parse(string(res))
	if err != nil {
		return nil, fmt.Errorf("template parse error: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	if err = t.Execute(buf, store); err != nil {
		return nil, fmt.Errorf("store template executor: %w", err)
	}

	return buf.Bytes(), nil
}
