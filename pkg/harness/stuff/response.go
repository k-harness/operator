package stuff

import (
	"bytes"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/util/jsonpath"
)

var (
	ErrNoKey       = errors.New("provided key not exists in result")
	ErrBadJsonPath = errors.New("bad json path formula")
	ErrBodyNil     = errors.New("body is nil")
)

// Response JSONPath exstractor
// info: https://kubernetes.io/docs/reference/kubectl/jsonpath/
type Response struct {
	Code string
	Body []byte
}

// GetKeyValue look up key in our json representation
// every time perform JSON marshaling. This is too slow!!!
func (a *Response) GetKeyValue(match string) (string, error) {
	if a.Body == nil {
		return "", ErrBodyNil
	}

	if match[0] != '{' {
		return a.GetKeyValue(fmt.Sprintf("{%s}", match))
	}

	var tmp interface{}
	if err := json.Unmarshal(a.Body, &tmp); err != nil {
		return "", fmt.Errorf("can't unmarshal Body: %w", err)
	}

	j := jsonpath.New(match)
	if err := j.Parse(match); err != nil {
		return "", fmt.Errorf("can't parse json path[%w]", err)
	}

	buf := bytes.NewBuffer(nil)
	if err := j.Execute(buf, tmp); err != nil {
		return "", fmt.Errorf("jsonpath execute error[%s]: %w", err, ErrNoKey)
	}

	if buf.String() == match {
		return "", ErrBadJsonPath
	}

	return buf.String(), nil
}
