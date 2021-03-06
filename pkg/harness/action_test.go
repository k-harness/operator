package harness

import (
	"bytes"
	"testing"

	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/k-harness/operator/pkg/harness/stuff"
	"github.com/k-harness/operator/pkg/harness/variables"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/util/jsonpath"
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
				KV: map[string]string{"KEY": `{{.MSG}}.{{.NUM}}`},
			},
			[]byte(`{"KEY":"HELLO.123"}`),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			act := NewRequest(
				"name",
				&v1alpha1.Request{Body: test.body},
				variables.New(test.kv, nil),
			)

			act.vars.RequestTranslate(act.Request)
			res, err := stuff.ScenarioBody(&act.Body).Get()

			assert.NoError(t, err)
			assert.Equal(t, test.want, res)
		})
	}
}

func TestActionResult_GetKeyValue(t *testing.T) {
	type fields struct {
		Code string
		Body []byte
	}

	tests := []struct {
		name    string
		fields  fields
		key     string
		want    string
		wantErr error
	}{
		{
			"has json",
			fields{Body: []byte(`{"message": 1}`)},
			"{.message}",
			"1",
			nil,
		},
		{
			"bad json path",
			fields{Body: []byte(`{"message": 1}`)},
			".message2",
			"",
			stuff.ErrBadJsonPath,
		},
		{
			"bad path",
			fields{Body: []byte(`{"message": 1}`)},
			"{.QQQ}",
			"",
			stuff.ErrNoKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &stuff.Response{
				Code: tt.fields.Code,
				Body: tt.fields.Body,
			}

			got, err := a.GetKeyValue(tt.key)
			assert.True(t, (tt.wantErr == nil) == (err == nil), err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// https://github.com/kubernetes/kubernetes/blob/758c56cc85e554122602d233cb315a07dd0e6961/pkg/util/jsonpath/jsonpath_test.go
func TestJsonPath(t *testing.T) {
	jp := jsonpath.New("x")
	assert.NoError(t, jp.Parse("{.message}"))
	buf := bytes.NewBuffer(nil)

	err := jp.Execute(buf, map[string]interface{}{"message": 1})
	assert.NoError(t, err)
	assert.Equal(t, "1", buf.String())
}
