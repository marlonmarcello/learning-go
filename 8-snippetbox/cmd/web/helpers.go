package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

type rootTemplateData struct {
	CurrentYear int
	PageData    any
}

// The serverError helper writes a log entry at Error level (including the request method and URI as attributes), then sends a generic 500 Internal Server Error response to the user.
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	method := r.Method
	uri := r.URL.RequestURI()
	trace := string(debug.Stack())

	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)

	// http.StatusText returns a human friendly text representation of the http code, like 400 would be "bad request"
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding description to the user. We use this to send responses like 400 "Bad Request" when there's a problem with the request that the user sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data rootTemplateData) {

	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	// we are going to render the template into a buffer to catch any runtime erros that might happens, like, misformed data, wrong action blocks, etc
	buff := new(bytes.Buffer)

	// write to buffer instead of the response
	err := ts.ExecuteTemplate(buff, "root", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// okay, now that we are clear of template errors, we write the status
	w.WriteHeader(status)

	// and finally, write the content of the buffer to the writer
	buff.WriteTo(w)
}

func newTemplateData[T any](x T) rootTemplateData {
	return rootTemplateData{
		CurrentYear: time.Now().Year(),
		PageData:    x,
	}
}

// Middleware that stops navigating to index routes without an index.html
func noIndexing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
