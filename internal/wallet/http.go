package wallet

import (
	"net/http"
	"time"

	"github.com/Ayan7020/go-pay/internal/platform/httpx"
	"github.com/google/uuid"
)

type Handler struct {
	srv *Service
}

type (
	walletResponse struct {
		ID        uuid.UUID  `json:"id"`
		OwnerID   *uuid.UUID `json:"owner_id"`
		Currency  string     `json:"currency"`
		CreatedAt time.Time  `json:"created_at"`
	}
)

func NewHandler(srv *Service) *Handler {
	return &Handler{
		srv: srv,
	}
}

func toWallet(a Account) walletResponse {
	return walletResponse{ID: a.ID, OwnerID: a.OwnerID, Currency: a.Currency, CreatedAt: a.CreatedAt}
}

func (h *Handler) createWallet(w http.ResponseWriter, r *http.Request) error {
	var req struct {
		OwnerID  string `json:"owner_id"`
		Currency string `json:"currency"`
	}

	if err := httpx.DecodeJson(w, r, &req); err != nil {
		return err
	}

	acc, err := h.srv.CreateWallet(r.Context(), req.OwnerID, req.Currency)
	if err != nil {
		return err
	}

	w.Header().Set("Location", "/v1/wallets/"+acc.ID.String())
	httpx.WriteJson(w, r, http.StatusCreated, toWallet(acc))
	return nil
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("POST /v1/wallets", httpx.HandlerFunc(h.createWallet))
}
