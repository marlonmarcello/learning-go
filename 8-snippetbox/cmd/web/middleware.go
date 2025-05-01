package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/justinas/nosurf"
)

/*
This is similar to the Alice package used on the course.

I wanted to try to implement something myself and I did a few iterations of this, and ended up learning quite a bit about Interface vs Concrete Types and how that's handled as parameters.
*/
type Middleware func(http.Handler) http.Handler

type MiddlewareChain struct {
	handlers []Middleware
}

func (m *MiddlewareChain) Append(handlers ...Middleware) {
	if m.handlers != nil {
		m.handlers = append(m.handlers, handlers...)
	} else {
		m.handlers = handlers
	}
}

func (m *MiddlewareChain) Then(next http.Handler) http.Handler {
	total := len(m.handlers)

	if total == 0 {
		return next
	}

	for i := total - 1; i >= 0; i-- {
		next = m.handlers[i](next)
	}

	return next
}

func (m *MiddlewareChain) ThenFunc(h http.HandlerFunc) http.Handler {
	return m.Then(h)
}

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

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			start  = time.Now()
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)

		wrapped := &wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		app.logger.Info("Request", "status", wrapped.statusCode, "ip", ip, "proto", proto, "method", method, "uri", uri, "time", time.Since(start))
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

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the user is not authenticated, redirect them to the login page and
		// return from the middleware chain so that no subsequent handlers in
		// the chain are executed.
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		// Otherwise set the "Cache-Control: no-store" header so that pages
		// require authentication are not stored in the users browser cache (or
		// other intermediary cache).
		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

// Create a NoSurf middleware function which uses a customized CSRF cookie with
// the Secure, Path and HttpOnly attributes set.
func (app *application) noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the authenticatedUserID value from the session using the
		// GetInt() method. This will return the zero value for an int (0) if no
		// "authenticatedUserID" value is in the session -- in which case we
		// call the next handler in the chain as normal and return.
		id := app.sessionManager.GetInt(r.Context(), authenticatedUserIDSessionKey)
		if id == 0 {
			next.ServeHTTP(w, r)
			return
		}

		// Otherwise, we check to see if a user with that ID exists in our database.
		exists, err := app.users.Exists(id)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		// If a matching user is found, we know that the request is
		// coming from an authenticated user who exists in our database. We
		// create a new copy of the request (with an isAuthenticatedContextKey
		// value of true in the request context) and assign it to r.
		if exists {
			ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}
