package httpx_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/Ayan7020/go-pay/internal/platform/httpx"
	"github.com/Ayan7020/go-pay/internal/platform/logging"
)

func TestRequestId(t *testing.T) {
	tests := []struct {
		name, header string
		err          bool
	}{
		{"generated when missing", "", false},
		{"kept when valid", "checkout-7f3a", false},
		{"replaced when invalid", "bad id\nlevel=ERROR", true},
	}

	for _, tt := range tests {
		var rId string
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(httpx.RequestIDHeader, tt.header)
		h := httpx.RequestId(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rId = httpx.RequestIDFrom(r.Context())
		}))
		h.ServeHTTP(rec, req)

		if rId == "" {
			t.Errorf("Invalid request_id= %v", rId)
		}

		if tt.header != "" {
			if rId != tt.header && !tt.err {
				t.Errorf("request_id= %v, want %v", rId, tt.header)
			}
			if tt.err && tt.header == rId {
				t.Errorf("on Error request_id should be different")
			}
		}

	}

}

func newLogger(buf *bytes.Buffer) *slog.Logger {
	return logging.New(buf, logging.Options{
		Service: "Test",
		Format:  "json",
	})
}

func TestAccessLog(t *testing.T) {
	var buf *bytes.Buffer
	httpx.AccessLog(newLogger(buf))
}

func TestChain(t *testing.T) {
	var order []string
	middleware := func(name string) httpx.Middleware {
		return func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				h.ServeHTTP(w, r)
			})
		}
	}
	hand := httpx.Chain(http.NotFoundHandler(), middleware("First"), middleware("Second"), middleware("Third"))
	hand.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	if want := []string{"First", "Second", "Third"}; !slices.Equal(want, order) {
		t.Errorf("order = %v, want %v", order, want)
	}
}
