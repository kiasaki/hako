package main

import (
	"encoding/base64"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
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
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusInternalServerError)
	if err := t.ExecuteTemplate(w, "error", err.Error()); err != nil {
		w.Write([]byte("500 - Oops! An error occured"))
		w.Write([]byte("\n\n"))
		w.Write([]byte(err.Error()))
	}
}

func fileSignedURL(file *HakoFile) (string, error) {
	userPrefix := base64.RawURLEncoding.EncodeToString([]byte(file.Owner))
	filePath := filepath.Join(userPrefix, filepath.Clean(file.Path))
	return storage.SignedURL(
		os.Getenv("GOOGLE_BUCKET_ID"),
		filePath,
		&storage.SignedURLOptions{
			GoogleAccessID: jwtConfig.Email,
			PrivateKey:     jwtConfig.PrivateKey,
			Method:         "GET",
			Expires:        time.Now().Add(5 * time.Minute),
		},
	)
}

func bytesString(s uint64) string {
	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	base := float64(1000)
	if s < 10 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(math.Log(float64(s)) / math.Log(base))
	suffix := sizes[int(e)]
	val := math.Floor(float64(s)/math.Pow(base, e)*10+0.5) / 10
	f := "%.0f %s"
	if val < 10 {
		f = "%.1f %s"
	}

	return fmt.Sprintf(f, val, suffix)
}
