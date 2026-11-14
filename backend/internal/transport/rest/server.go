package rest

import (
	"fmt"
	"net/http"
	"time"
)

func BuildServer(host string, port string, router http.Handler) *http.Server {

	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", host, port),
		Handler:           router,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       8 * time.Second,
	}

	return server
}
