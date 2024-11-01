package main

import (
	"log"
	"net/http"
)

type api struct {
	addr string
}

func (s *api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// very very crude method and path matching
	// kinda like a little router
	switch r.Method {
	case http.MethodGet:
		switch r.URL.Path {
		case "/":
			w.Write([]byte("index page"))
			return
		case "/users":
			w.Write([]byte("users page"))
			return
		default:
			w.Write([]byte("404 page"))
		}
	default:
		w.Write([]byte("method not allowed"))
		return
	}
}

func main() {
	// implements the handle interface
	app := &api{addr: ":8080"}

	srv := &http.Server{
		Addr:    app.addr,
		Handler: app,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
