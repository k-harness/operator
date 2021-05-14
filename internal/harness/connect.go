package harness

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/k-harness/operator/api/v1alpha1/models/action"
)

var (
	ErrNoConnectionData = errors.New("no connection data")
)

type connect struct {
	*v1alpha1.Connect
}

func (c *connect) Call(ctx context.Context, body []byte) (*ActionResult, error) {
	if c.GRPC != nil {
		res, err := NewGRPCRequest(c.GRPC).Call(ctx, body)
		if err != nil {
			return nil, fmt.Errorf("grpc call: %w", err)
		}

		return res, nil
	}

	if c.HTTP != nil {
		res, err := NewHttpRequest(c.HTTP).Call(ctx, body)
		if err != nil {
			return nil, fmt.Errorf("http call: %w", err)
		}

		return res, nil
	}

	return nil, ErrNoConnectionData
}

func get(in *action.HTTP) (*ActionResult, error) {
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

	req, err := http.NewRequest(in.Method, uri.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("http init request %q erorr :%w", uri.String(), err)
	}

	hResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http action execute %q error %w", uri.String(), err)
	}

	res, err := io.ReadAll(hResp.Body)
	if err != nil {
		return nil, fmt.Errorf("http action read result error :%w", err)
	}

	return &ActionResult{Body: res, Code: hResp.Status}, nil
}
