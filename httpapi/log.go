package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Debug will pass debugging information to the client if true
var Debug = false

type statusWriter struct {
	http.ResponseWriter
	Status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.Status = code
	w.ResponseWriter.WriteHeader(code)
}

type logData struct {
	Action   string        `json:"action"`
	ActionID string        `json:"action_id,omitempty"`
	User     string        `json:"user,omitempty"`
	Result   int           `json:"result"`
	Error    string        `json:"error,omitempty"`
	Time     time.Time     `json:"time"`
	Duration time.Duration `json:"duration"`
}

func withLogging(action string, output io.Writer, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := &logData{Action: action}

		if id, ok := mux.Vars(r)["id"]; ok {
			l.ActionID = id
		}

		t := time.Now()

		sw := &statusWriter{ResponseWriter: w}
		ctx := context.WithValue(r.Context(), contextKeyLogData, l)
		next.ServeHTTP(sw, r.WithContext(ctx))

		l.Result = sw.Status
		l.Time = time.Now()
		l.Duration = l.Time.Sub(t)

		j, err := json.Marshal(l)
		if err != nil {
			log.Println("Unable to marshal JSON:", err)
		}
		_, err = fmt.Fprintln(output, string(j))
		if err != nil {
			log.Println("Unable to output log:", err)
		}
	})
}
