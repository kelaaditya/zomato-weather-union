package middlewares

import (
	"fmt"
	"net/http"
)

// recover panic, log panic message, and continue middleware chain
func (middleware *Middleware) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// set HTTP connection close
				w.Header().Set("Connection", "close")

				// http method type
				var HTTPMethod string = r.Method
				// request URL
				var requestURI string = r.RequestURI

				// log request parameters that generated error
				middleware.Logger.Error(
					fmt.Sprintf("error: %s", err),
					"method",
					HTTPMethod,
					"uri",
					requestURI,
				)

				// send error response
				http.Error(
					w,
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
