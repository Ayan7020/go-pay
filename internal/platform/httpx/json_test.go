package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Ayan7020/go-pay/internal/platform/httpx"
)

func TestDecodeJson(t *testing.T) {
	test_cases := []struct {
		name    string
		body    string
		status  int
		message string
	}{
		{"valid", `{"amount": 10}`, 200, ""},
		{"empty", ``, 400, "body must not be empty"},
		{"malformed", `{"amount": 5`, 400, "malformed JSON"},
		{"wrong type", `{"amount": "500"}`, 400, `field "amount" must be of type int64`},
		{"unknown field", `{"amount": 500, "amout": 5}`, 400, `unknown field "amout"`},
		{"two values", `{"amount": 1}{"amount": 2}`, 400, "body must contain a single JSON value"},
		{"too large", `{"amount": 1, "pad": "` + strings.Repeat("x", 2<<20) + `"}`, 413, "body must be at most 1048576 bytes"},
	}
	for _, test_case := range test_cases {
		t.Run(test_case.name, func(t *testing.T) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var v struct {
					Amount int64 `json:"amount"`
				}
				if err := httpx.DecodeJson(w, r, &v); err != nil {
					httpx.Error(w, r, err)
					return
				}

				httpx.WriteJson(w, r, http.StatusOK, v)
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/", strings.NewReader(test_case.body))
			h(rec, req)

			if rec.Code != test_case.status {
				t.Fatalf("status = %d, want %d", rec.Code, test_case.status)
			}

			if test_case.status != 200 {
				if got := decodeError(rec, t).Error.Message; got != test_case.message {
					t.Errorf("message = %q, want %q", got, test_case.message)
				}
			}
		})
	}
}
