package wallet_test

import (
	"os"
	"testing"

	"github.com/Ayan7020/go-pay/internal/platform/postgres"
	"github.com/Ayan7020/go-pay/internal/wallet"
	"github.com/google/uuid"
)

func newStore(t *testing.T) *wallet.Store {
	t.Helper()
	dsn := os.Getenv("WALLET_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("WALLET_TEST_DATABASE_URL not set, skipping integration test")
	}

	pool, err := postgres.NewPool(t.Context(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return wallet.NewStore(pool)
}

func newAccount(t *testing.T, s *wallet.Store, name string, side wallet.Direction) uuid.UUID {
	t.Helper()
	acc, err := s.CreateAccount(t.Context(), wallet.NewAccount{
		Name:          name,
		Currency:      "EUR",
		NormalBalance: side,
	})
	if err != nil {
		t.Fatalf("create account %s: %v", name, err)
	}

	return acc.ID
}

func TestPostTransaction(t *testing.T) {
	t.Parallel()
	s := newStore(t)
	newAccount(t, s, "bank", wallet.Debit)
}
