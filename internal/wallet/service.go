package wallet

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/Ayan7020/go-pay/internal/platform/apperr"
	"github.com/google/uuid"
)

type Service struct {
	store *Store
}

const SettlementAccount = "settlement"

var currencyCode = regexp.MustCompile(`^[A-Z]{3}$`)

func NewService(store *Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) CreateWallet(ctx context.Context, ownerId, currency string) (Account, error) {
	var valError apperr.ValidationError

	ownerUid, err := uuid.Parse(ownerId)
	if err != nil || ownerUid == uuid.Nil {
		valError.Add("ownerId", "must be an valid uuid")
	}

	if !currencyCode.MatchString(currency) {
		valError.Add("currency", "must be a three-letter ISO 4217 code, like USD")
	}

	if err := valError.Err(); err != nil {
		return Account{}, valError.Err()
	}

	_, err = s.store.GetSystemAccount(ctx, SettlementAccount, currency)
	if errors.Is(err, ErrAccountNotFound) {
		valError.Add("currency", "unsupported currency")
		return Account{}, valError.Err()
	}

	acc, err := s.store.CreateAccount(ctx, NewAccount{
		Kind:          KindWallet,
		OwnerID:       &ownerUid,
		Name:          currency + " wallet",
		Currency:      currency,
		NormalBalance: Credit,
	})
	if errors.Is(err, ErrAccountExists) {
		return Account{}, fmt.Errorf("%w: owner already has a wallet in %s", apperr.ErrConflict, currency)
	}

	return acc, err
}
