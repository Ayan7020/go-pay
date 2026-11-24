package httpx

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Ayan7020/go-pay/internal/platform/apperr"
)

type errorDetail struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

type errorBody struct {
	Error errorDetail `json:"error"`
}

func Classify(err error) (int, string) {
	var reqErr *requestErr
	var valErr *apperr.ValidationError
	switch {
	case errors.As(err, &reqErr):
		return reqErr.status, reqErr.code
	case errors.As(err, &valErr):
		return http.StatusUnprocessableEntity, "invalid_input"
	case errors.As(err, apperr.ErrNotFound):
		return http.StatusNotFound, "not_found"
	case errors.As(err, apperr.ErrConflict):
		return http.StatusConflict, "conflict"
	default:
		return http.StatusInternalServerError, "internal"
	}
}

func Error(w http.ResponseWriter, r *http.Request, err error) {
	status, code := Classify(err)
	d := errorDetail{Code: code, Message: err.Error()}
	var valErr *apperr.ValidationError
	if ok := errors.As(err, &valErr); ok {
		d.Fields = valErr.Fields
	}
	if status >= 500 {
		slog.ErrorContext(r.Context(), "request failed", "err", err)
		d.Message = "internal server error"
	}

	writeError(w, r, status, d)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, d errorDetail) {
	d.RequestID = RequestIDFrom(r.Context())
	body, _ := json.Marshal(errorBody{Error: d})
	writeBody(w, status, body)
}
