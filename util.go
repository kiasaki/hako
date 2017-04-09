package main

import (
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

func match(method, pathPrefix string, handler func(w http.ResponseWriter, r *http.Request)) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Handle index as a direct match, not prefix (so than 404 works)
			if pathPrefix == "/" {
				if r.URL.Path == "/" {
					handler(w, r)
				} else {
					h.ServeHTTP(w, r)
				}
				return
			}

			if r.Method == method && strings.HasPrefix(r.URL.Path, pathPrefix) {
				handler(w, r)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

func renderTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	if err := t.ExecuteTemplate(w, name, data); err != nil {
		sendError(w, r, err)
	}
}

func sendError(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	if err := t.ExecuteTemplate(w, "error", err); err != nil {
		w.Write([]byte("Internal Server Error"))
		if os.Getenv("DEBUG") != "" {
			w.Write([]byte("\n\n"))
			w.Write([]byte(err.Error()))
		}
	}
}

func generateSignedUrl(invoiceId string) (string, error) {
	return storage.SignedURL(
		os.Getenv("GOOGLE_BUCKET_ID"),
		invoiceId+".pdf",
		&storage.SignedURLOptions{
			GoogleAccessID: jwtConfig.Email,
			PrivateKey:     jwtConfig.PrivateKey,
			Method:         "GET",
			Expires:        time.Now().Add(15 * time.Minute),
		},
	)
}
