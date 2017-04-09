package main

import (
	"net/http"
	"strings"
)

// GetCookie retrieves and verifies the cookie value.
// from: https://github.com/lgtmco/lgtm/blob/master/shared/httputil/httputil.go
func GetCookie(r *http.Request, name string) (value string) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return
	}
	value = cookie.Value
	return
}

// SetCookie writes the cookie value.
// from: https://github.com/lgtmco/lgtm/blob/master/shared/httputil/httputil.go
func SetCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   r.URL.Host,
		HttpOnly: true,
		Secure:   IsHttps(r),
		MaxAge:   2147483647, // the cooke value (token) is responsible for expiration
	}

	http.SetCookie(w, &cookie)
}

// DelCookie deletes a cookie.
// from: https://github.com/lgtmco/lgtm/blob/master/shared/httputil/httputil.go
func DelCookie(w http.ResponseWriter, r *http.Request, name string) {
	cookie := http.Cookie{
		Name:   name,
		Value:  "deleted",
		Path:   "/",
		Domain: r.URL.Host,
		MaxAge: -1,
	}

	http.SetCookie(w, &cookie)
}

// IsHttps is a helper function that evaluates the http.Request
// and returns True if the Request uses HTTPS. It is able to detect,
// using the X-Forwarded-Proto, if the original request was HTTPS and
// routed through a reverse proxy with SSL termination.
// from: https://github.com/lgtmco/lgtm/blob/master/shared/httputil/httputil.go
func IsHttps(r *http.Request) bool {
	switch {
	case r.URL.Scheme == "https":
		return true
	case r.TLS != nil:
		return true
	case strings.HasPrefix(r.Proto, "HTTPS"):
		return true
	case r.Header.Get("X-Forwarded-Proto") == "https":
		return true
	default:
		return false
	}
}
