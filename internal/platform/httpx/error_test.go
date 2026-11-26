package httpx_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/Ayan7020/go-pay/internal/platform/apperr"
	"github.com/Ayan7020/go-pay/internal/platform/httpx"
)

type errorResponse struct {
	Error struct {
		Code      string            `json:"code"`
		Message   string            `json:"message"`
		RequestID string            `json:"request_id"`
		Fields    map[string]string `json:"fields"`
	} `json:"error"`
}

func decodeError(responseRecorder *httptest.ResponseRecorder, t *testing.T) errorResponse {
	t.Helper()
	var body errorResponse
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body %q: %v", responseRecorder.Body, err)
	}
	return body
}

func TestError(t *testing.T) {
	valErr := &apperr.ValidationError{}
	valErr.Add("name", "atleast have 6 characters")
	valErr.Add("amount", "must be positive")

	test_err_cases := []struct {
		name    string
		err     error
		status  int
		code    string
		message string
	}{
		{"Not found", fmt.Errorf("account test: %w", apperr.ErrNotFound), 404, "not_found", "account test: not found"},
		{"Conflict", fmt.Errorf("wallet exists: %w", apperr.ErrConflict), 409, "conflict", "wallet exists: conflict"},
		{"Invalid", valErr.Err(), 422, "invalid_input", "invalid input: amount, name"},
		{"Internal", errors.New("dial tcp 10.0.0.7:5432: refused"), 500, "internal", "internal server error"},
	}

	for _, tec := range test_err_cases {
		t.Run(tec.name, func(t *testing.T) {
			resrec := httptest.NewRecorder()
			httpx.Error(resrec, httptest.NewRequest("GET", "/", nil), tec.err)

			if resrec.Code != tec.status {
				t.Errorf("status = %d, want %d", resrec.Code, tec.status)
			}

			body := decodeError(resrec, t)
			if body.Error.Code != tec.code || body.Error.Message != tec.message {
				t.Errorf("got %q %q, want %q %q", body.Error.Code, body.Error.Message, tec.code, tec.message)
			}
		})
	}
}
