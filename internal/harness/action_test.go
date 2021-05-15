package harness

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/k-harness/operator/api/v1alpha1"
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
				KV: map[string]v1alpha1.Any{"KEY": `{{.MSG}}.{{.NUM}}`},
			},
			[]byte(`{"KEY":"HELLO.123"}`),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			act := NewAction("name", v1alpha1.Action{Request: v1alpha1.Request{Body: test.body}}, test.kv)

			res, err := act.GetBody()
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
			"message",
			"",
			ErrBadJsonPath,
		},
		{
			"bad path",
			fields{Body: []byte(`{"message": 1}`)},
			"{.QQQ}",
			"",
			ErrNoKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &ActionResult{
				Code: tt.fields.Code,
				Body: tt.fields.Body,
			}
			got, err := a.GetKeyValue(tt.key)
			fmt.Println(got)

			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr))
			}

			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("GetKeyValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetKeyValue() got = %v(%[1]T), want %v(%[2]T)", got, tt.want)
			}
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
