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
	Res    interface{}
	Status int
}

func CreateMockServer(fx Fixture) (net.Listener, *http.Server, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, nil, fmt.Errorf("listen erorr :%w", err)
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if res, err := io.ReadAll(req.Body); err == nil {
			log.Println("BODY:", string(res))
		}

		if err := json.NewEncoder(w).Encode(fx.Res); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}

		w.WriteHeader(fx.Status)
	})

	srv := &http.Server{
		Addr:    l.Addr().String(),
		Handler: h,
	}

	go func() {
		if err := srv.Serve(l); err != nil {
			log.Println(err.Error())
		}
	}()

	return l, srv, nil
}
