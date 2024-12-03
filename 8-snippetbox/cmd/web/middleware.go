package main

import (
	"fmt"
	"net/http"
)

/*
  In any middleware handler, code which comes before next.ServeHTTP() will be executed on the way down the chain, and any code after next.ServeHTTP() — or in a deferred function — will be executed on the way back up.

  Another thing to mention is that if you call return in your middleware function before you call next.ServeHTTP(), then the chain will stop being executed and control will flow back upstream.

  As an example, a common use-case for early returns is authentication middleware which only allows execution of the chain to continue if a particular check is passed.
*/

func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		/*
		  Important: You must make sure that your response header map contains all the headers you want before you call w.WriteHeader() or w.Write(). Any changes you make to the response header map after calling w.WriteHeader() or w.Write() will have no effect on the headers that the user receives.
		*/
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")

		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		w.Header().Set("Server", "Go")

		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)

		app.logger.Info("Request", "ip", ip, "proto", proto, "method", method, "uri", uri)

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// create a deferred function (which will always be run in the event of a panic as Go unwinds the stack)
		defer func() {
			// use the builtin recover function to check if there has been a panic or not
			if err := recover(); err != nil {
				// inform clients that the connection will be closed, as Go will do that on panics
				w.Header().Set("Connection", "close")
				// gracefully returns a 500 error
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
