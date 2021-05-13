package models

import (
	"testing"

	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestAction_GetBody(t *testing.T) {
	tests := []struct {
		name string
		kv   map[string]string
		body v1alpha1.Body
		want []byte
	}{
		{
			"OK JSON",
			map[string]string{"MSG": "HELLO"},
			v1alpha1.Body{
				Row: `{"KEY":"{{.MSG}}"}`,
			},
			[]byte(`{"KEY":"HELLO"}`),
		},
		{
			"OK BYTE",
			map[string]string{"MSG": "HELLO"},
			v1alpha1.Body{
				Byte: []byte(`{"KEY":"{{.MSG}}"}`),
			},
			[]byte(`{"KEY":"HELLO"}`),
		},
		{
			"OK KV",
			map[string]string{"MSG": "HELLO", "NUM": "123"},
			v1alpha1.Body{
				KV: map[string]v1alpha1.Any{"KEY": `{{.MSG}}.{{.NUM}}`},
			},
			[]byte(`{"KEY":"HELLO.123"}`),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			act := NewAction("name", v1alpha1.Action{Request: v1alpha1.Request{Body: test.body}})

			res, err := act.GetBody(test.kv)
			assert.NoError(t, err)
			assert.Equal(t, test.want, res)
		})
	}
}
