package stuff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		jsonPath string
		res      string
		err      bool
	}{
		{
			"OK",
			Response{Body: []byte(`{"test":"1"}`)},
			"{.test}",
			"1",
			false,
		},
		{
			"$ root",
			Response{Body: []byte(`{"test":"1"}`)},
			".test",
			"1",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := test.response.GetKeyValue(test.jsonPath)
			assert.True(t, test.err == (err != nil), err)
			assert.Equal(t, test.res, res)
		})
	}
}
