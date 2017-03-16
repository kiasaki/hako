package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/option"
)

type H map[string]interface{}

var t *template.Template
var jwtConfig *jwt.Config
var ctx context.Context
var bucket *storage.BucketHandle

func init() {
	// Load templates
	t = template.Must(template.ParseGlob("templates/*"))
	t = template.Must(t.ParseGlob("public/*"))

	// Setup Google Cloud Datastore client
	ctx = context.Background()
	conf, err := google.JWTConfigFromJSON(
		[]byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")), storage.ScopeFullControl)
	if err != nil {
		log.Fatal(err)
	}
	jwtConfig = conf
	client, err := storage.NewClient(
		ctx,
		option.WithTokenSource(conf.TokenSource(ctx)),
	)
	if err != nil {
		log.Fatal(err)
	}

	bucket = client.Bucket(os.Getenv("GOOGLE_BUCKET_ID"))
}

func main() {
	http.HandleFunc("/", indexHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("started listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
func renderTemplate(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	if err := t.ExecuteTemplate(w, name, data); err != nil {
		errorHandler(w, r, err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r, "index", nil)

}

func errorHandler(w http.ResponseWriter, r *http.Request, err error) {
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
