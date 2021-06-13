package harness

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/k-harness/operator/api/v1alpha1/models/action"
	"github.com/k-harness/operator/pkg/harness/stuff"
)

type httpRequest struct {
	*action.HTTP
}

func NewHttpRequest(in *action.HTTP) RequestInterface {
	return &httpRequest{HTTP: in}
}

func (in *httpRequest) Call(ctx context.Context, request *v1alpha1.Request) (*stuff.Response, error) {
	uri, err := url.Parse(in.Addr)
	if err != nil {
		return nil, fmt.Errorf("bad http address: %w", err)
	}

	if in.Path != nil {
		uri.Path = path.Join(uri.Path, *in.Path)
	}

	if in.Query != nil {
		q := make(url.Values)
		for k, v := range in.Query {
			q.Add(k, v)
		}

		uri.RawQuery = q.Encode()
	}

	var hResp *http.Response

	if in.HTTP.Form {
		if len(request.Body.KV) == 0 {
			return nil, fmt.Errorf("http post form require only KV body")
		}

		val := make(url.Values)
		for k, v := range request.Body.KV {
			val[k] = []string{v}
		}

		hResp, err = http.DefaultClient.PostForm(uri.String(), val)
	} else {
		var body []byte
		body, err = stuff.ScenarioBody(&request.Body).Get()
		if err != nil {
			return nil, fmt.Errorf("action can't exstract body: %w", err)
		}

		var req *http.Request
		req, err = http.NewRequest(in.Method, uri.String(), bytes.NewBuffer(body))
		if err != nil {
			return nil, fmt.Errorf("http init request %q erorr :%w", uri.String(), err)
		}

		for key, val := range request.Header {
			req.Header.Add(key, val)
		}

		if request.Body.Type == "json" {
			req.Header.Add("content-type", "application/json")
		}

		hResp, err = new(http.Client).Do(req)
	}

	if err != nil {
		return nil, fmt.Errorf("http action execute %q error %w", uri.String(), err)
	}

	res, err := io.ReadAll(hResp.Body)
	if err != nil {
		return nil, fmt.Errorf("http action read result error :%w", err)
	}

	return &stuff.Response{Body: res, Code: fmt.Sprintf("%d", hResp.StatusCode)}, nil
}
