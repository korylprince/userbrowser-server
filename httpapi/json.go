package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"reflect"
)

type returnHandlerFunc func(*http.Request) (int, interface{})

type jsonResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
	Debug       string `json:"debug,omitempty"`
}

func withJSONResponse(next returnHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code, body := next(r)

		if err, ok := body.(error); ok || body == nil {
			resp := jsonResponse{Code: code, Description: http.StatusText(code)}
			body = resp
			if err != nil {
				(r.Context().Value(contextKeyLogData)).(*logData).Error = err.Error()
				if er, ok := err.(*errResponse); ok {
					body = er
				} else if Debug {
					resp.Debug = err.Error()
				}
			}
		}

		w.Header().Set(headerContentType, "application/json")
		w.WriteHeader(code)

		e := json.NewEncoder(w)
		err := e.Encode(body)

		if err != nil {
			log.Println("Error writing JSON response:", err)
		}
	})
}

func jsonRequest(r *http.Request, v interface{}) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get(headerContentType))
	if err != nil {
		return fmt.Errorf("Could not parse Content-Type: %v", err)
	}

	if mediaType != mediaTypeJSON {
		return errors.New("Content-Type not application/json")
	}

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("Unable to parse request body to %s: %v", reflect.TypeOf(v).Elem().Name(), err)
	}

	return nil
}
