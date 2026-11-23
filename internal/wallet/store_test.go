package wallet_test

import (
	"errors"
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

func balance(t *testing.T, id uuid.UUID) int64 {
	t.Parallel()
	s := newStore(t)
	ab, err := s.Balance(t.Context(), id)
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	return ab.Balance
}

func TestPostTransaction(t *testing.T) {
	t.Parallel()
	s := newStore(t)
	newAccount(t, s, "bank", wallet.Debit)
}

func TestAccountNotFound(t *testing.T) {
	t.Parallel()
	s := newStore(t)

	_, err := s.Account(t.Context(), uuid.New())
	if !errors.Is(err, wallet.ErrAccountNotFound) {
		t.Fatalf("got error %v, want ErrAccountNotFound", err)
	}
}
