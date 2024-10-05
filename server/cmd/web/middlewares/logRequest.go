package middlewares

import (
	"net/http"
)

// log all incoming requests to the server
func (middleware *Middleware) LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// request parameters for logging
		var ip string = r.RemoteAddr
		var proto string = r.Proto
		var method string = r.Method
		var uri string = r.URL.RequestURI()

		// log each request via the application logger
		middleware.Logger.Info(
			"received request",
			"ip",
			ip,
			"proto",
			proto,
			"method",
			method,
			"uri",
			uri,
		)

		// call the next-in-line
		next.ServeHTTP(w, r)
	})
}
