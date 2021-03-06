package httpexec

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"k8s.io/apimachinery/pkg/util/json"
)

type Fixture struct {
	Addr   string
	Res    interface{}
	Status int

	// accepted request
	RequestAccepted struct {
		BodyRow []byte
		BodyMap map[string]interface{}
		// save only first header in slice of key
		Headers map[string]string
	}
}

func CreateMockServer(fx *Fixture) (net.Listener, *http.Server, error) {
	if fx.Addr == "" {
		fx.Addr = ":0"
	}

	l, err := net.Listen("tcp", fx.Addr)
	if err != nil {
		return nil, nil, fmt.Errorf("listen erorr :%w", err)
	}

	x := http.NewServeMux()
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// read request
		if res, err := io.ReadAll(req.Body); err == nil {
			fx.RequestAccepted.BodyRow = res
			_ = json.Unmarshal(res, &fx.RequestAccepted.BodyMap)
		}

		fx.RequestAccepted.Headers = make(map[string]string)
		for key, val := range req.Header {
			fx.RequestAccepted.Headers[key] = val[0]
		}

		// send response
		if err := json.NewEncoder(w).Encode(fx.Res); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}

		w.WriteHeader(fx.Status)
	})

	x.Handle("/", h)
	x.Handle("/echo", http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if res, err := io.ReadAll(req.Body); err == nil {
			fx.RequestAccepted.BodyRow = res
			_, _ = writer.Write(res)
			return
		}

		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(err.Error()))
	}))

	srv := &http.Server{Addr: l.Addr().String(), Handler: x}

	go func() {
		if err := srv.Serve(l); err != nil {
			log.Println(err.Error())
		}
	}()

	return l, srv, nil
}
