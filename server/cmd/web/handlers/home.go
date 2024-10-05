package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func (handler *Handler) Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// http method type
		var HTTPMethod string = r.Method
		// request URL
		var requestURI string = r.RequestURI

		// get calculations for display on the map
		calculations, err :=
			handler.
				Models.
				Calculation.GetCalculationsTemperatureWithStationDetails(
				context.Background(),
			)
		if err != nil {
			// log error
			handler.Logger.Error(
				"no data found when fetching for calculations",
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
		// if no data fetched
		if len(calculations) == 0 {
			// log error
			handler.Logger.Error(
				"no data found when fetching for calculations",
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

		// convert data to JSON
		// as the template JS does not have object transfer from the backend
		calculationJSONBytes, err := json.Marshal(calculations)
		if err != nil {
			// log error
			handler.Logger.Error(
				"error in converting data to JSON",
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

		// convert slice of bytes JSON to string JSON (utf-8)
		dataForTemplate := string(calculationJSONBytes)

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
		err = HTMLTemplate.ExecuteTemplate(
			HTMLTemplateBuffer,
			"base",
			dataForTemplate,
		)
		if err != nil {
			// log error
			handler.Logger.Error(
				"error in executing home page template",
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
				"error in writing bytes to the writer in home page template",
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
