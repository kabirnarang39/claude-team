package httpx

import (
	"net/http"
	"strings"
)

// HTTPSOnlyMiddleware rejects requests that are not served over HTTPS with
// a 400 Bad Request response and sets a Strict-Transport-Security header.
//
// In devMode, requests from localhost/127.0.0.1/::1 are allowed over HTTP
// to support local development without TLS.
func HTTPSOnlyMiddleware(devMode bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set HSTS on all responses (including rejections) so browsers upgrade future requests.
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

			if isHTTPS(r) {
				next.ServeHTTP(w, r)
				return
			}

			if devMode && isLocalhost(r) {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, "HTTPS is required", http.StatusBadRequest)
		})
	}
}

// isHTTPS returns true if the request was made over TLS, either directly or
// via a reverse proxy that sets X-Forwarded-Proto.
func isHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); strings.EqualFold(proto, "https") {
		return true
	}
	return false
}

// isLocalhost returns true if the request originates from a loopback address.
func isLocalhost(r *http.Request) bool {
	host := r.Host
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}
