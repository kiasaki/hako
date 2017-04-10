package main

import (
	"errors"
	"net/http"
	"path/filepath"
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
	http.Redirect(w, r, "/", 302)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = "/v/"
	handleView(w, r)
}

func loadCurrentFileAndFolder(email, filePath string) (*HakoFile, []*HakoFile, error) {
	file := &HakoFile{
		Owner: email,
		Path:  filePath,
	}
	if filePath == "" {
		file.Path = "."
	}
	err := storageGet(file)
	if err != nil {
		return nil, nil, err
	}

	folderFiles, err := storageList(file.Owner, file.Folder())
	if err != nil {
		return nil, nil, err
	}

	return file, folderFiles, nil
}

func handleNew(w http.ResponseWriter, r *http.Request) {
	email := authEmailFromContext(r.Context())
	folderPath := r.URL.Path[len("/v/"):]
	file, folderFiles, err := loadCurrentFileAndFolder(email, folderPath)
	if err != nil {
		sendError(w, r, err)
		return
	}

	renderTemplate(w, r, "new", H{
		"file":        file,
		"folderFiles": folderFiles,
	})
}

func handleNewSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		sendError(w, r, errors.New("Error parsing form"))
		return
	}
	fileName := r.FormValue("name")
	if fileName == "" {
		sendError(w, r, errors.New("Invalid file name"))
		return
	}

	email := authEmailFromContext(r.Context())
	folder := HakoFile{Path: r.URL.Path[len("/n/"):]}
	filePath := filepath.Clean(filepath.Join(folder.Folder(), fileName))
	if fileName[len(fileName)-1] == '/' {
		filePath += ".folder"
	}
	file := &HakoFile{
		Owner:    email,
		Path:     filePath,
		Contents: []byte{},
	}
	err := storagePut(file)
	if err != nil {
		sendError(w, r, err)
		return
	}

	http.Redirect(w, r, "/v/"+file.Path, 302)
}

func handleView(w http.ResponseWriter, r *http.Request) {
	email := authEmailFromContext(r.Context())
	file, folderFiles, err := loadCurrentFileAndFolder(email, r.URL.Path[len("/v/"):])
	if err != nil {
		sendError(w, r, err)
		return
	}

	renderTemplate(w, r, "view", H{
		"file":        file,
		"folderFiles": folderFiles,
	})
}

func handleEdit(w http.ResponseWriter, r *http.Request) {
	email := authEmailFromContext(r.Context())
	file, folderFiles, err := loadCurrentFileAndFolder(email, r.URL.Path[len("/e/"):])
	if err != nil {
		sendError(w, r, err)
		return
	}

	renderTemplate(w, r, "edit", H{
		"file":        file,
		"folderFiles": folderFiles,
	})
}

func handleEditSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		sendError(w, r, errors.New("Error parsing form"))
		return
	}

	file := &HakoFile{
		Owner:    authEmailFromContext(r.Context()),
		Path:     r.URL.Path[len("/e/"):],
		Contents: []byte(r.FormValue("contents")),
	}
	if err := storagePut(file); err != nil {
		sendError(w, r, err)
		return
	}

	http.Redirect(w, r, "/v/"+file.Path, 302)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "view", nil)
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "not-found", nil)
}
