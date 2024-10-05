package handlers

import (
	"bytes"
	"net/http"
)

func (handler *Handler) Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// http method type
		var HTTPMethod string = r.Method
		// request URL
		var requestURI string = r.RequestURI

		// get home page HTML template from cache
		HTMLTemplate, ok := handler.TemplateCache["home"]
		if !ok {
			// log error
			handler.Logger.Error(
				"home page template file not found in html cache",
				"method",
				HTTPMethod,
				"uri",
				requestURI,
			)
			// error with built-in status
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}

		// initialize new buffer
		HTMLTemplateBuffer := new(bytes.Buffer)

		// execute the HTML template
		err := HTMLTemplate.ExecuteTemplate(HTMLTemplateBuffer, "base", nil)
		if err != nil {
			// log error
			handler.Logger.Error(
				"home page template file not found in html cache",
				"method",
				HTTPMethod,
				"uri",
				requestURI,
			)
			// error with built-in status
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}

		// if no error in the HTML template execution
		// write the buffer to w
		_, err = w.Write(HTMLTemplateBuffer.Bytes())
		if err != nil {
			// log error
			handler.Logger.Error(
				"home page template file not found in html cache",
				"method",
				HTTPMethod,
				"uri",
				requestURI,
			)
			// error with built-in status
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}
	}
}
