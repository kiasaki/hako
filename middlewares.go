package main

import (
	"context"
	"errors"
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
		// On error validating token, go to sign in
		sessionToken := GetCookie(r, authCookieName)
		userEmail, err := validateSessionToken(sessionToken)
		if err != nil {
			log.Printf("auth: got error validating token: %v\n", err)
			http.Redirect(w, r, "/signin", 302)
			return
		}

		// Refresh session token for 2 weeks
		sessionToken, err = createSessionToken(userEmail)
		if err != nil {
			sendError(w, r, errors.New("Error creating a session token"))
			return
		}
		SetCookie(w, r, authCookieName, sessionToken)

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
