package main

import (
	"path/filepath"
	"text/template"
	"time"

	"github.com/marlonmarcello/learning-go/8-snippetbox/internal/models"
)

type snippetViewTemplateData struct {
	Snippet models.Snippet
}

type homeTemplateData struct {
	Snippets []models.Snippet
}

// template functions can take as many arguments as they need but MUST return a single value, UNLESS the second value is an error
func humanDate(t time.Time) string {
	return t.Format("02 jan 2006 at 15:04")
}

// this will act as a lookup between the names of our functions
var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// this will give us a slice of all the filepaths for our application page templates
	pages, err := filepath.Glob("./ui/html/pages/*.tmpl.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		// extract the filename, like 'home.tmpl.html' from the fill filepath
		name := filepath.Base(page)

		// parse root template
		// the template.FuncMap must be registered with the template set before we can call parse, that means we need to use template.New to create an empty set, and register with Funcs
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/templates/root.tmpl.html")
		if err != nil {
			return nil, err
		}

		// parse partials
		ts, err = ts.ParseGlob("./ui/html/partials/*.tmpl.html")
		if err != nil {
			return nil, err
		}

		// parse page template
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// store to cache with page template name
		cache[name] = ts
	}

	return cache, nil
}
