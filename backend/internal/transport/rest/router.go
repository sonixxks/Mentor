package rest

import (
	"fmt"
	"net/http"
)

func NewRouter(handlers []ServiceHandler, ver string) *http.ServeMux {

	rootMux := http.NewServeMux()
	mux := http.NewServeMux()

	for _, handler := range handlers {
		handler.Register(mux)
	}

	prefix := fmt.Sprintf("/api/%s", ver)

	rootMux.Handle(prefix+"/", http.StripPrefix(prefix, mux))

	return rootMux
}

type ServiceHandler interface {
	Register(*http.ServeMux)
}
