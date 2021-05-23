package harness

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/k-harness/operator/api/v1alpha1/models/action"
	executor2 "github.com/k-harness/operator/pkg/executor"
)

type httpRequest struct {
	*action.HTTP
}

func NewHttpRequest(in *action.HTTP) RequestInterface {
	return &httpRequest{HTTP: in}
}

func (in *httpRequest) Call(ctx context.Context, request *executor2.Request) (*ActionResult, error) {
	uri, err := url.Parse(in.Addr)
	if err != nil {
		return nil, fmt.Errorf("bad http address: %w", err)
	}

	if in.Path != nil {
		uri.Path = path.Join(uri.Path, *in.Path)
	}

	if in.Query != nil {
		uri.RawQuery = *in.Query
	}

	req, err := http.NewRequest(in.Method, uri.String(), bytes.NewBuffer(request.Body))
	if err != nil {
		return nil, fmt.Errorf("http init request %q erorr :%w", uri.String(), err)
	}

	for key, val := range request.Header {
		req.Header.Add(key, val)
	}

	if request.Type == "json" {
		req.Header.Add("content-type", "application/json")
	}

	hResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http action execute %q error %w", uri.String(), err)
	}

	res, err := io.ReadAll(hResp.Body)
	if err != nil {
		return nil, fmt.Errorf("http action read result error :%w", err)
	}

	return &ActionResult{Body: res, Code: fmt.Sprintf("%d", hResp.StatusCode)}, nil
}
