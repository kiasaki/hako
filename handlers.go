package main

import (
	"encoding/base64"
	"errors"
	"io"
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

func loadCurrentFileAndFolder(email, filePath string) (*HakoFile, *HakoFile, []*HakoFile, error) {
	file := &HakoFile{
		Owner: email,
		Path:  filePath,
	}
	if !file.IsFolder() {
		err := storageGet(file)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	currentFolder := file
	if !file.IsFolder() {
		currentFolder = &HakoFile{
			Owner: email,
			Path:  file.ParentPath(),
		}
	}
	folderFiles, err := storageList(currentFolder)
	if err != nil {
		return nil, nil, nil, err
	}

	return file, currentFolder, folderFiles, nil
}

func handleFetch(w http.ResponseWriter, r *http.Request) {
	file := &HakoFile{
		Owner: authEmailFromContext(r.Context()),
		Path:  r.URL.Path[len("/f/"):],
	}
	err := storageGet(file)
	if err != nil {
		sendError(w, r, err)
		return
	}
	if file.Type() == "text" {
		w.Header().Set("Content-Type", "text/plain")
	} else if file.Type() == "markdown" {
		w.Header().Set("Content-Type", "text/plain")
	} else if file.Type() == "image" {
		switch file.Ext() {
		case "png":
			w.Header().Set("Content-Type", "image/png")
		case "gif":
			w.Header().Set("Content-Type", "image/gif")
		case "svg":
			w.Header().Set("Content-Type", "image/svg+xml")
		case "jpg":
			w.Header().Set("Content-Type", "image/jpeg")
		case "jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
		default:
			w.Header().Set("Content-Type", "image/jpeg")
		}
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename="+file.Name())
	}
	w.Write(file.Contents)
}

func handleNew(w http.ResponseWriter, r *http.Request) {
	email := authEmailFromContext(r.Context())
	folderPath := r.URL.Path[len("/v/"):]
	file, folder, folderFiles, err := loadCurrentFileAndFolder(email, folderPath)
	if err != nil {
		sendError(w, r, err)
		return
	}

	renderTemplate(w, r, "new", H{
		"file":        file,
		"folder":      folder,
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

	file := &HakoFile{
		Owner:    authEmailFromContext(r.Context()),
		Path:     filepath.Clean(filepath.Join(r.URL.Path[len("/n/"):], fileName)),
		Contents: []byte{},
	}
	err := storagePut(file)
	if err != nil {
		sendError(w, r, err)
		return
	}

	http.Redirect(w, r, "/v/"+file.Path, 302)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	email := authEmailFromContext(r.Context())
	folderPath := r.URL.Path[len("/v/"):]
	file, folder, folderFiles, err := loadCurrentFileAndFolder(email, folderPath)
	if err != nil {
		sendError(w, r, err)
		return
	}

	renderTemplate(w, r, "upload", H{
		"file":        file,
		"folder":      folder,
		"folderFiles": folderFiles,
	})
}

func handleUploadSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		sendError(w, r, errors.New("Error parsing form"))
		return
	}

	f, handler, err := r.FormFile("file")
	if err != nil {
		sendError(w, r, errors.New("Error reading file"))
		return
	}
	defer f.Close()

	fileName := handler.Filename
	if filepath.Ext(fileName) == "" {
		fileName += ".txt" // Ensure we have an extension
	}
	file := &HakoFile{
		Owner:    authEmailFromContext(r.Context()),
		Path:     filepath.Clean(filepath.Join(r.URL.Path[len("/n/"):], fileName)),
		Contents: []byte{},
	}

	// RAW Google Cloud call here for efficient copying
	userPrefix := base64.RawURLEncoding.EncodeToString([]byte(file.Owner))
	filePath := filepath.Join(userPrefix, filepath.Clean(file.Path))
	objHandle := bucket.Object(filePath)
	wc := objHandle.NewWriter(ctx)

	if _, err := io.Copy(wc, f); err != nil {
		sendError(w, r, err)
		return
	}
	if err = wc.Close(); err != nil {
		sendError(w, r, err)
		return
	}

	http.Redirect(w, r, "/v/"+file.Path, 302)
}

func handleView(w http.ResponseWriter, r *http.Request) {
	email := authEmailFromContext(r.Context())
	file, folder, folderFiles, err := loadCurrentFileAndFolder(email, r.URL.Path[len("/v/"):])
	if err != nil {
		sendError(w, r, err)
		return
	}

	renderTemplate(w, r, "view", H{
		"file":        file,
		"folder":      folder,
		"folderFiles": folderFiles,
	})
}

func handleEdit(w http.ResponseWriter, r *http.Request) {
	email := authEmailFromContext(r.Context())
	file, folder, folderFiles, err := loadCurrentFileAndFolder(email, r.URL.Path[len("/e/"):])
	if err != nil {
		sendError(w, r, err)
		return
	}

	renderTemplate(w, r, "edit", H{
		"file":        file,
		"folder":      folder,
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
	email := authEmailFromContext(r.Context())
	file, _, folderFiles, err := loadCurrentFileAndFolder(email, r.URL.Path[len("/d/"):])
	if err != nil {
		sendError(w, r, err)
		return
	}
	if file.IsFolder() && len(folderFiles) > 0 {
		sendError(w, r, errors.New("Can't delete a folder that still contains files."))
		return
	}
	err = storageDel(file)
	if err != nil {
		sendError(w, r, err)
		return
	}
	if file.ParentPath() == "." {
		http.Redirect(w, r, "/v/", 302)
	} else {
		http.Redirect(w, r, "/v/"+file.ParentPath(), 302)
	}
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "not-found", nil)
}
