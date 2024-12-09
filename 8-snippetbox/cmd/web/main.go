package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"text/template"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marlonmarcello/learning-go/8-snippetbox/internal/models"
)

// application struct will hold application-wide dependencies for the web application.
type application struct {
	logger        *slog.Logger
	snippets      *models.SnippetModel
	templateCache map[string]*template.Template
}

/*
  Lesson 02.10 has an explanation on Handlers that is really important.
  Couple bullet points:
  - Handler is a simple interface, all it needs is a ServeHTTP method
  - mux.HandleFunc is just syntatic sugar that transforms a function into a Handler by calling the passed function as the ServeHTTP method so we don't have to declare a struct just to conform to the interface
  - notice how we use mux.Handle for the static/ file server, it's because http.FileServer is already a Handler
  - you can chain handlers, which is the default http mux do
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

	// defer the db connection close so it runs before the end of main()
	defer db.Close()

	// initialize templates cache
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// initialize application with all dependencies
	app := &application{
		logger:        logger,
		snippets:      &models.SnippetModel{DB: db, CTX: ctx},
		templateCache: templateCache,
	}

	logger.Info("starting server", "addr", *addr)

	/*
	  You can think of a Go web application as a chain of ServeHTTP() methods being called one after another.

	  When our server receives a new HTTP request it calls the servemux’s ServeHTTP() method. This looks up the relevant handler based on the request method and URL path, and in turn calls that handler’s ServeHTTP() method.
	*/
	err = http.ListenAndServe(*addr, app.routes())
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
