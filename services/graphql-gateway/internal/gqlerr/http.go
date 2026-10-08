package gqlerr

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// HTTPStatus wraps the GraphQL handler so a response with a FORBIDDEN error
// gets HTTP status 403 instead of 200. The body is unchanged: clients still
// read extensions.code, and any data the other fields returned.
func HTTPStatus(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &recorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		status := rec.status
		if status == http.StatusOK && hasCode(rec.body.Bytes(), CodeForbidden) {
			status = http.StatusForbidden
		}
		w.WriteHeader(status)
		_, _ = w.Write(rec.body.Bytes())
	})
}

// recorder holds the response back until its errors are known. Headers (like
// Set-Cookie) go straight to the real response.
type recorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (r *recorder) WriteHeader(status int)      { r.status = status }
func (r *recorder) Write(b []byte) (int, error) { return r.body.Write(b) }

func hasCode(body []byte, code string) bool {
	var response struct {
		Errors []struct {
			Extensions struct {
				Code string `json:"code"`
			} `json:"extensions"`
		} `json:"errors"`
	}
	if json.Unmarshal(body, &response) != nil {
		return false
	}
	for _, e := range response.Errors {
		if e.Extensions.Code == code {
			return true
		}
	}
	return false
}
