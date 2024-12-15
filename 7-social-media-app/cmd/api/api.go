package main

import (
	"log"
	"net/http"
	"time"
)

type api struct {
	config apiConfig
}

type apiConfig struct {
	addr string
}

func (app *api) mount() *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("GET /healthcheck", app.healthCheckGetHandler)

	// Sub all routes to a v1 prefix
	v1Router := http.NewServeMux()
	v1Router.Handle("/v1/", http.StripPrefix("/v1", router))

	return v1Router
}

func (app *api) serve(mux *http.ServeMux) error {
	basicMiddleware := CreateMiddlewareStack(
		RecoverMiddleware,
		LoggingMiddleware,
	)

	handler := basicMiddleware(mux)

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      handler,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Printf("Server started at %s", app.config.addr)

	return srv.ListenAndServe()
}
