package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type requestErr struct {
	status int
	code   string
	msg    string
}

func WriteJson(w http.ResponseWriter, r *http.Request, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		Error(w, r, fmt.Errorf("encode response: %w", err))
		return
	}
	writeBody(w, status, body)
}

func writeBody(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(append(body, '\n'))
}
