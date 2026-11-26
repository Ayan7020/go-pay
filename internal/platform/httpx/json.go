package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxBodyBytes = 1 << 20

type requestErr struct {
	status int
	code   string
	msg    string
}

func (e *requestErr) Error() string {
	return e.msg
}

func WriteJson(w http.ResponseWriter, r *http.Request, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		Error(w, r, fmt.Errorf("encode response: %w", err))
		return
	}
	writeBody(w, status, body)
}

func badRequest(msg string) error {
	return &requestErr{http.StatusBadRequest, "bad_request", msg}
}

func DecodeJson(w http.ResponseWriter, r *http.Request, dst any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return decodeError(err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return badRequest("body must contain a single JSON value")
	}
	return nil
}

func decodeError(err error) error {
	var sizeError *http.MaxBytesError
	var typeErr *json.UnmarshalTypeError
	var syntaxErr *json.SyntaxError
	switch {
	case errors.Is(err, io.EOF):
		return badRequest("body must not be empty")
	case errors.As(err, &syntaxErr):
		return badRequest(fmt.Sprintf("malformed JSON at byte %d", syntaxErr.Offset))
	case errors.Is(err, io.ErrUnexpectedEOF):
		return badRequest("malformed JSON")
	case errors.As(err, &typeErr):
		return badRequest(fmt.Sprintf("field %q must be of type %s", typeErr.Field, typeErr.Type))
	case strings.HasPrefix(err.Error(), "json: unknown field"):
		return badRequest(strings.TrimPrefix(err.Error(), "json: "))
	case errors.As(err, &sizeError):
		msg := fmt.Sprintf("body must be at most %d bytes", sizeError.Limit)
		return &requestErr{http.StatusRequestEntityTooLarge, "body_to_large", msg}
	default:
		return badRequest("invalid JSON body")
	}
}

func writeBody(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(append(body, '\n'))
}
