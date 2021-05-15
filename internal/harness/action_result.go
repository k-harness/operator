package harness

import (
	"bytes"
	"fmt"

	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/util/jsonpath"
)

type ActionResult struct {
	Code string
	Body []byte
}

// GetKeyValue look up key in our json representation
// every time perform JSON marshaling. This is too slow!!!
func (a *ActionResult) GetKeyValue(jsonPath string) (string, error) {
	if a.Body == nil {
		return "", fmt.Errorf("body is nil")
	}

	var tmp interface{}
	if err := json.Unmarshal(a.Body, &tmp); err != nil {
		return "", fmt.Errorf("can't unmarshal body: %w", err)
	}

	j := jsonpath.New(jsonPath)
	if err := j.Parse(jsonPath); err != nil {
		return "", fmt.Errorf("can't parse json path[%w]", err)
	}

	buf := bytes.NewBuffer(nil)
	if err := j.Execute(buf, tmp); err != nil {
		return "", fmt.Errorf("jsonpath execute error[%s]: %w", err, ErrNoKey)
	}

	if buf.String() == jsonPath {
		return "", ErrBadJsonPath
	}

	return buf.String(), nil
}
