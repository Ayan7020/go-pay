package wallet

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ayan7020/go-pay/internal/wallet/walletdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Account        = walletdb.Account
	AccountBalance = walletdb.AccountBalance
	NewAccount     = walletdb.CreateAccountParams
	Direction      = walletdb.EntryDirection
)

var ErrAccountNotFound = errors.New("wallet: account not found")

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

func (s *Store) Account(ctx context.Context, id uuid.UUID) (Account, error) {
	acc, err := s.q.GetAccount(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrAccountNotFound
	}
	if err != nil {
		return Account{}, fmt.Errorf("get account %s: %w", id, err)
	}

	return acc, nil
}

func (s *Store) Balance(ctx context.Context, id uuid.UUID) (AccountBalance, error) {
	ab, err := s.q.GetAccountBalance(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return AccountBalance{}, ErrAccountNotFound
	}
	if err != nil {
		return AccountBalance{}, fmt.Errorf("get balance %s: %w", id, err)
	}
	return ab, nil
}

