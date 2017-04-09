package main

import (
	"log"
	"net/http"
	"time"
)

func middlewareRequireAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO
		handler.ServeHTTP(w, r)
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
