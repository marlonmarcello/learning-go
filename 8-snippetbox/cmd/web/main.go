package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marlonmarcello/learning-go/8-snippetbox/internal/models"
)

// application struct will hold application-wide dependencies for the web application.
type application struct {
	logger         *slog.Logger
	snippets       *models.SnippetModel
	users          *models.UserModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

/*
  Lesson 02.10 has an explanation on Handlers that is really important.
  Couple bullet points:
  - Handler is a simple interface, all it needs is a ServeHTTP method
  - mux.HandleFunc is just syntatic sugar that transforms a function into a Handler by calling the passed function as the ServeHTTP method so we don't have to declare a struct just to conform to the interface
  - notice how we use mux.Handle for the static/ file server, it's because http.FileServer is already a Handler
  - you can chain handlers, which is the default http mux do
*/

/*
Advanced article on interfaces.
https://jordanorelli.com/post/32665860244/how-to-use-interfaces-in-go
- interfaces area also a type, so you can create an array of interfaces for example but the implementations are different

The Interfaces section of the Effective Go docs is good:
https://go.dev/doc/effective_go#interfaces
-
*/

func main() {
	addr := flag.String("addr", ":8080", "HTTP network address")

	// default db connection string
	dsn := flag.String("dsn", "postgres://web:pass@localhost/snippetbox", "Postgres data source name")

	/*
	  Important to note that for flags defined with flag.Bool(), omitting a value when starting the application is the same as writing -flag=true.
	*/
	debug := flag.Bool("debug", true, "Shows debug messages")

	// this needs to be called BEFORE using any flags as it will store them in their correct pointers
	flag.Parse()

	// by default debug logs are silenced
	loggerLevel := slog.LevelInfo
	addSource := false
	if *debug {
		loggerLevel = slog.LevelDebug

		// this will ensure we add the filename and line number of the calling source
		addSource = true
	}

	// custom structured logger, outputs to standard out and uses default options
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     loggerLevel,
		AddSource: addSource,
	}))

	// open pool of db connections
	ctx := context.Background()
	db, err := openDb(ctx, *dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	// This ensures that cookies are only sent over HTTPS
	sessionManager.Cookie.Secure = true

	// defer the db connection close so it runs before the end of main()
	defer db.Close()

	// initialize templates cache
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	formDecoder := form.NewDecoder()

	// initialize application with all dependencies
	app := &application{
		logger:         logger,
		snippets:       &models.SnippetModel{DB: db, CTX: ctx},
		users:          &models.UserModel{DB: db, CTX: ctx},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	tlsConfig := &tls.Config{
		// Go supports a few elliptic curves, but as of Go 1.23 only tls.CurveP256 and tls.X25519 have assembly implementations. The others are very CPU intensive, so omitting them helps ensure that our server will remain performant under heavy loads.
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},

		/*
		   Lesson 9.05 explains more about cypher suites and how Go utilizes them with TLS connections.
		   For backwards compatibility reasons, to support TLS1.0-1.2, Go does allow and uses by default some older cypher suites that are vulnerable to attacks
		   All cyphers used by Go for TLS1.3 are safe by default and always used
		   This list is a list of cypher suites that work with TLS1.2 and are considered safe.
		   Reccomended configuration: https://wiki.mozilla.org/Security/Server_Side_TLS

		   Marlon's opinion: The drawback here is that we are excluding TLS1.0 but if a user has such an older browser, they shouldn't be on the internet.
		*/
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
		},
	}

	srv := &http.Server{
		Addr:    *addr,
		Handler: app.routes(),
		/*
			It’s important to be aware that Go’s http.Server may write its own log entries relating to things like unrecovered panics, or problems accepting or writing to HTTP connections.

			By default, it writes these entries using the standard logger — which means they will be written to the standard error stream (instead of standard out like our other log entries), and they won’t be in the same format as the rest of our nice structured log entries.
		*/
		ErrorLog:  slog.NewLogLogger(logger.Handler(), loggerLevel),
		TLSConfig: tlsConfig,
		/*
		  Timeouts to improve resiliency
		  All three of these timeouts — IdleTimeout, ReadTimeout and WriteTimeout — are server-wide settings which act on the underlying connection and apply to all requests irrespective of their handler or URL.

		  There are really good explanations for all of these on Lesson 9.06
		*/
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("starting server", "addr", srv.Addr)

	/*
	   You can think of a Go web application as a chain of ServeHTTP() methods being called one after another.

	   When our server receives a new HTTP request it calls the servemux’s ServeHTTP() method. This looks up the relevant handler based on the request method and URL path, and in turn calls that handler’s ServeHTTP() method.

	   https://fideloper.com/golang-http-handlers
	*/

	// [NOTE] Using HTTPS - Go’s will automatically upgrade the connection to use HTTP/2 if the client supports it.
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	logger.Error(err.Error())
	os.Exit(1)
}

func openDb(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	err = dbpool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return dbpool, nil
}
