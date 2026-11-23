package wallet

import (
	"context"
	"fmt"

	"github.com/Ayan7020/go-pay/internal/wallet/walletdb"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Account    = walletdb.Account
	NewAccount = walletdb.CreateAccountParams
	Direction  = walletdb.EntryDirection
)

// var ErrAccountNotFound = errors.New("wallet: account not found")

const (
	Debit  = walletdb.EntryDirectionDebit
	Credit = walletdb.EntryDirectionCredit
)

type Store struct {
	pool *pgxpool.Pool
	q    *walletdb.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, q: walletdb.New(pool)}
}

func (s *Store) CreateAccount(ctx context.Context, a NewAccount) (Account, error) {
	acc, err := s.q.CreateAccount(ctx, a)
	if err != nil {
		return Account{}, fmt.Errorf("create account: %w", err)
	}

	return acc, nil
}
