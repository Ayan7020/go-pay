package wallet_test

import (
	"errors"
	"os"
	"testing"

	"github.com/Ayan7020/go-pay/internal/platform/postgres"
	"github.com/Ayan7020/go-pay/internal/wallet"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func newStore(t *testing.T) *wallet.Store {
	t.Helper()
	dsn := os.Getenv("WALLET_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("WALLET_TEST_DATABASE_URL not set, skipping integration test")
	}

	pool, err := postgres.NewPool(t.Context(), dsn,2)
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

func balance(t *testing.T, s *wallet.Store, id uuid.UUID) int64 {
	t.Helper()
	ab, err := s.Balance(t.Context(), id)
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	return ab.Balance
}

func TestPostTransaction(t *testing.T) {
	t.Parallel()
	s := newStore(t)
	bank := newAccount(t, s, "bank", wallet.Debit)
	kalento := newAccount(t, s, "Kalento", wallet.Credit)

	tx, err := s.PostTransaction(t.Context(), wallet.NewTransaction{
		Description: "Kalento deposits €50",
		Posting: []wallet.Posting{
			{AccountID: bank, Direction: wallet.Debit, Amount: 5000, Currency: "EUR"},
			{AccountID: kalento, Direction: wallet.Credit, Amount: 5000, Currency: "EUR"},
		},
	})
	if err != nil {
		t.Fatalf("post transaction: %v", err)
	}

	if got := balance(t, s, bank); got != 5000 {
		t.Errorf("bank balance = %d, want 5000", got)
	}

	if got := balance(t, s, kalento); got != 5000 {
		t.Errorf("bank balance = %d, want 5000", got)
	}

	entries, err := s.Entries(t.Context(), kalento, 0, 10)
	if err != nil {
		t.Fatalf("entries: %v", err)
	}

	if len(entries) != 1 || entries[0].TransactionID != tx.ID {
		t.Errorf("kalento has %d entries, want 1 in transaction %s", len(entries), tx.ID)
	}
}

func TestPostTransactionRejectsUnbalanced(t *testing.T) {
	t.Parallel()
	s := newStore(t)
	ayan := newAccount(t, s, "ayan", wallet.Debit)
	kalento := newAccount(t, s, "Kalento", wallet.Credit)

	_, err := s.PostTransaction(t.Context(), wallet.NewTransaction{
		Description: "Kalento deposits €50",
		Posting: []wallet.Posting{
			{AccountID: ayan, Direction: wallet.Debit, Amount: 5000, Currency: "EUR"},
			{AccountID: kalento, Direction: wallet.Credit, Amount: 4000, Currency: "EUR"},
		},
	})

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23514" {
		t.Fatalf("got error %v, want a check_violation (23514) from Postgres", err)
	}

	t.Logf("rejected by Postgres: %s", pgErr.Message)

	if got := balance(t, s, ayan); got != 0 {
		t.Errorf("ayan balance = %d, want 0", got)
	}
	if got := balance(t, s, kalento); got != 0 {
		t.Errorf("kalento balance = %d, want 0", got)
	}
}

func TestAccountNotFound(t *testing.T) {
	t.Parallel()
	s := newStore(t)

	_, err := s.Account(t.Context(), uuid.New())
	if !errors.Is(err, wallet.ErrAccountNotFound) {
		t.Fatalf("got error %v, want ErrAccountNotFound", err)
	}
}
