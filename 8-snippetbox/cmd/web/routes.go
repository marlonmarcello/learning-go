package main

import "net/http"

func (app *application) routes() http.Handler {
	/*
	  Really important shit to know: incoming HTTP requests are served in their own go routines, which makes go really fast but can cause race conditions if you are not careful, read more:
	  https://www.alexedwards.net/blog/understanding-mutexes
	*/
	mux := http.NewServeMux()

	// Path is relative to project root
	staticFileServer := http.FileServer(http.Dir("./ui/static/"))

	/*
	  Register the file server as the handler for all URL paths that start with /static/ and for anything that matches, we strip the "/static" prefix before the request reacher the file server
	*/
	mux.Handle("GET /static/", http.StripPrefix("/static", noIndexing(staticFileServer)))

	// Middleware stack for our main pages
	dynamicStack := MiddlewareChain{
		handlers: []Middleware{
			app.sessionManager.LoadAndSave,
		},
	}

	/*
	  When a route pattern ends with a trailing slash — like "/" or "/static/" — it is known as a subtree path pattern. Subtree path patterns are matched (and the corresponding handler called) whenever the start of a request URL path matches the subtree path.

	  If it helps your understanding, you can think of subtree paths as acting a bit like they have a wildcard at the end, like "/**" or "/static/**".

	  This helps explain why the "/" route pattern acts like a catch-all. The pattern essentially means match a single slash, followed by anything (or nothing at all).

	  To prevent subtree path patterns from acting like they have a wildcard at the end, you can append the spec
	*/
	mux.Handle("GET /{$}", dynamicStack.ThenFunc(app.home))

	/*
	  When a pattern doesn’t have a trailing slash, it will only be matched (and the corresponding handler called) when the request URL path exactly matches the pattern in full.
	*/
	mux.Handle("GET /snippet/view/{id}", dynamicStack.ThenFunc(app.snippetView))
	mux.Handle("GET /snippet/create", dynamicStack.ThenFunc(app.snippetCreate))
	mux.Handle("POST /snippet/create", dynamicStack.ThenFunc(app.snippetCreatePost))

	// Authentication routes
	mux.Handle("GET /user/signup", dynamicStack.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamicStack.ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/login", dynamicStack.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamicStack.ThenFunc(app.userLoginPost))
	mux.Handle("POST /user/logout", dynamicStack.ThenFunc(app.userLogoutPost))

	/*
	  Pass the servemux as the 'next' parameter to the commonHeaders middleware.
	  Because commonHeaders is just a function, and the function returns a http.Handler we don't need to do anything else.

	  It’s important to know that when the last handler in the chain returns, control is passed back up the chain in the reverse direction. So when our code is being executed the flow of control actually looks like this:

	  recoverPanic → logRequest → commonHeaders → servemux → application handler → servemux → commonHeaders → logRequest → recoverPanic
	*/
	standardStack := MiddlewareChain{
		handlers: []Middleware{
			app.recoverPanic,
			app.logRequest,
			commonHeaders,
		}}

	return standardStack.Then(mux)
}
