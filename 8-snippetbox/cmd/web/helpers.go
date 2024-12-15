package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-playground/form/v4"
)

type rootTemplateData struct {
	CurrentYear  int
	FlashMessage string
	PageData     any
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

	/*
	  If you don’t call w.WriteHeader() explicitly, then the first call to w.Write() will automatically send a 200 status code to the user. So, if you want to send a non-200 status code, you must call w.WriteHeader() before any call to w.Write().
	*/
	// okay, now that we are clear of template errors, we write the status
	w.WriteHeader(status)

	// and finally, write the content of the buffer to the writer
	buff.WriteTo(w)
}

func (app *application) newTemplateData(r *http.Request, x any) rootTemplateData {
	return rootTemplateData{
		CurrentYear:  time.Now().Year(),
		FlashMessage: app.sessionManager.PopString(r.Context(), "flash"),
		PageData:     x,
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

func (app *application) decodePostForm(r *http.Request, destination any) error {
	/*
	   First we call r.ParseForm() which adds any data in POST request bodies to the r.PostForm map. This also works in the same way for PUT and PATCH requests.

	   ParseForm is safe to be called multiple times per request
	*/
	err := r.ParseForm()
	if err != nil {
		return err
	}

	/*
	  Chapter 7.2 has more information on how to use different types  of data with PostForm.Get(), like multiple checkboxes and multipart form data
	*/

	err = app.formDecoder.Decode(destination, r.PostForm)
	if err != nil {
		/*
		   If we try to use an invalid target destination, the Decode() method will return an error with the type *form.InvalidDecoderError.We use errors.As() to check for this and raise a panic rather than returning the error.
		*/
		var invalidDecoder *form.InvalidDecoderError

		if errors.As(err, &invalidDecoder) {
			panic(err)
		}

		return err
	}

	return nil
}
