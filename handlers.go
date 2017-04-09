package main

import (
	"errors"
	"net/http"
	"regexp"
)

var emailRegexp = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

func handleSignin(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "signin", nil)
}

func handleSigninSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		sendError(w, r, errors.New("Error parsing form"))
		return
	}
	email := r.FormValue("email")
	if !emailRegexp.MatchString(email) {
		sendError(w, r, errors.New("Invalid email address provided"))
		return
	}
	if err := emailSigninLink(email); err != nil {
		sendError(w, r, errors.New("Error sending sign in link email"))
		return
	}
	renderTemplate(w, r, "signin-success", nil)
}

func handleSignout(w http.ResponseWriter, r *http.Request) {
	DelCookie(w, r, authCookieName)
	http.Redirect(w, r, "/signin", 302)
}

func handleSigninLink(w http.ResponseWriter, r *http.Request) {
	signinToken := r.URL.Query().Get("t")
	email, err := validateSigninToken(signinToken)
	if err != nil {
		sendError(w, r, errors.New("Invalid sign in token provided"))
		return
	}

	sessionToken, err := createSessionToken(email)
	if err != nil {
		sendError(w, r, errors.New("Error creating a session token"))
		return
	}

	SetCookie(w, r, authCookieName, sessionToken)
	http.Redirect(w, r, "/v/0", 302)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/v/0", 302)
}

func handleNew(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "index", nil)
}

func handleNewSubmit(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "index", nil)
}

func handleView(w http.ResponseWriter, r *http.Request) {
	email := authEmailFromContext(r.Context())
	renderTemplate(w, r, "index", email)
}

func handleEdit(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "index", nil)
}

func handleEditSubmit(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "index", nil)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "index", nil)
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "not-found", nil)
}
