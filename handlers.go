package main

import (
	"net/http"
)

func handleSignin(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "index", nil)
}

func handleSigninSubmit(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "index", nil)
}

func handleSignout(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "index", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "index", nil)
}

func handleNew(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "index", nil)
}

func handleNewSubmit(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "index", nil)
}

func handleView(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "index", nil)
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
