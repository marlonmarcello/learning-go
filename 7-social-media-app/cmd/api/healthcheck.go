package main

import "net/http"

func (app *api) healthCheckGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Healthy"))
}
