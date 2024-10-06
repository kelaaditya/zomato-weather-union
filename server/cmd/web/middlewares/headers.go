package middlewares

import (
	"net/http"
)

// set common headers
func (middleware *Middleware) CommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// set CSP
		w.Header().Set(
			"Content-Security-Policy",
			`
				default-src 'self';
				style-src 'self' fonts.googleapis.com https://unpkg.com;
				font-src fonts.gstatic.com;
				script-src 'self' 'nonce-J2B9AHS41HA8' https://unpkg.com;
				img-src 'self' https://tile.openstreetmap.org https://openstreetmap.org data:;
			`,
		)

		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		w.Header().Set("Server", "Go")

		// call the next-in-line
		next.ServeHTTP(w, r)
	})
}
