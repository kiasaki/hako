package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

type contextKey string

func authEmailFromContext(ctx context.Context) string {
	if v := ctx.Value(contextKey("email")); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func middlewareRequireAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionToken := GetCookie(r, authCookieName)
		userEmail, err := validateSessionToken(sessionToken)

		// On error validating token, go to sign in
		if err != nil {
			log.Printf("auth: got error validating token: %v\n", err)
			http.Redirect(w, r, "/signin", 302)
			return
		}

		// Else save user's email in context and go on
		ctx := context.WithValue(r.Context(), contextKey("email"), userEmail)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

func middlewareLogWithTiming(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(started time.Time) {
			timing := time.Since(started).Nanoseconds() / 1000.0
			log.Printf("%s: %s (%dus)\n", r.Method, r.RequestURI, timing)
		}(time.Now())
		handler.ServeHTTP(w, r)
	})
}
