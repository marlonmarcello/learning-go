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

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r, userSignupTemplateData{})
	app.render(w, r, http.StatusOK, "signup.tmpl.html", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var templateData userSignupTemplateData

	err := app.decodePostForm(r, &templateData)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	templateData.CheckField(validator.NotBlank(templateData.Name), "name", "This field cannot be blank")
	templateData.CheckField(validator.NotBlank(templateData.Email), "email", "This field cannot be blank")
	templateData.CheckField(validator.Matches(templateData.Email, validator.EmailRX), "email", "This field must be a valid email")
	templateData.CheckField(validator.NotBlank(templateData.Password), "password", "This field cannot be blank")
	templateData.CheckField(validator.MinChars(templateData.Password, 8), "password", "This field must be at least 8 characters")

	if !templateData.Valid() {
		data := app.newTemplateData(r, templateData)

		app.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl.html", data)

		return
	}

	_, err = app.users.Insert(templateData.Name, templateData.Email, templateData.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			templateData.AddFormError("email", "Email address already in use")

			data := app.newTemplateData(r, templateData)

			app.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
		} else {
			app.serverError(w, r, err)
		}

		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r, userLoginTemplateData{})
	app.render(w, r, http.StatusOK, "login.tmpl.html", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var templateData userLoginTemplateData

	err := app.decodePostForm(r, &templateData)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	templateData.CheckField(validator.NotBlank(templateData.Email), "email", "This field cannot be blank")
	templateData.CheckField(validator.Matches(templateData.Email, validator.EmailRX), "email", "This field must be a valid email")
	templateData.CheckField(validator.NotBlank(templateData.Password), "password", "This field cannot be blank")

	if !templateData.Valid() {
		data := app.newTemplateData(r, templateData)
		app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl.html", data)
		return
	}

	id, err := app.users.Authenticate(templateData.Email, templateData.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			templateData.AddNonFieldError("Email or password is incorrect")
			data := app.newTemplateData(r, templateData)
			app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl.html", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	// Use the RenewToken() method on the current session to change the session
	// ID. It's good practice to generate a new session ID when the
	// authentication state or privilege levels changes for the user (e.g. login
	// and logout operations).
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Add the ID of the current user to the session, so that they are now
	// 'logged in'.
	app.sessionManager.Put(r.Context(), "authenticatedUserId", id)

	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserId")

	app.sessionManager.Put(r.Context(), "flash", "You logged out successfully")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
