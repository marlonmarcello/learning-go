package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/marlonmarcello/learning-go/8-snippetbox/internal/models"
	"github.com/marlonmarcello/learning-go/8-snippetbox/internal/validator"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r, homeTemplateData{
		Snippets: snippets,
	})

	app.render(w, r, http.StatusOK, "home.tmpl.html", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	// strvonv.Atoi tries to parse a string into an integer base 10
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}

		return
	}

	data := app.newTemplateData(r, snippetViewTemplateData{
		Snippet: snippet,
	})

	app.render(w, r, http.StatusOK, "view.tmpl.html", data)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r, snippetCreateTemplateData{
		Expires: 1,
	})
	app.render(w, r, http.StatusOK, "create.tmpl.html", data)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {

	// we'll pass this to the decoder to populate the fields
	var templateData snippetCreateTemplateData

	err := app.decodePostForm(r, &templateData)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	templateData.CheckField(validator.NotBlank(templateData.Title), "title", "This field cannot be blank")
	templateData.CheckField(validator.MaxChars(templateData.Title, 100), "title", "This field cannot be more than 100 characters long")

	/*
		  Not necessary anymore, the decoder already handles type convertion

			if !validator.NotBlank(rawExpire) {
				templateData.CheckField(true, "expires", "You must select at least one expiry option")
			} else {
				// PostForm.Get() always returns the form data as a *string*.
				// However, we're expecting our expires value to be a number, and want to represent it in our Go code as an integer. So we need to manually convert the form data using strconv.Atoi(), and we send a 400 in case that fails
				expires, err = strconv.Atoi(r.PostForm.Get("expires"))
				if err != nil {
					app.clientError(w, http.StatusBadRequest)
					return
				}
			}
	*/

	templateData.CheckField(validator.NotBlank(templateData.Content), "content", "This field cannot be blank")

	templateData.CheckField(validator.PermittedValue(templateData.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !templateData.Valid() {
		data := app.newTemplateData(r, templateData)

		app.render(w, r, http.StatusUnprocessableEntity, "create.tmpl.html", data)

		return
	}

	id, err := app.snippets.Insert(templateData.Title, templateData.Content, time.Now().AddDate(0, 0, templateData.Expires))
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
