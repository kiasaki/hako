package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"

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

var appBaseURL = os.Getenv("APP_BASE_URL")
var jwtSecret = []byte(os.Getenv("APP_JWT_SECRET"))
var sendgridApiKey = os.Getenv("SENDGRID_API_KEY")

const homeFolderName = "Home.folder"
const authCookieName = "hako_session"

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
	handler := NewChain(
		middlewareLogWithTiming,
		match("GET", "/signin", handleSignin),
		match("POST", "/signin", handleSigninSubmit),
		match("GET", "/signout", handleSignout),
		match("GET", "/sl", handleSigninLink),
		middlewareRequireAuth,
		match("GET", "/", handleIndex),
		match("GET", "/f/", handleFetch),
		match("GET", "/n/", handleNew),
		match("POST", "/n/", handleNewSubmit),
		match("GET", "/v/", handleView),
		match("GET", "/e/", handleEdit),
		match("POST", "/e/", handleEditSubmit),
		match("GET", "/d/", handleDelete),
	).Then(http.HandlerFunc(handleNotFound))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("started listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
