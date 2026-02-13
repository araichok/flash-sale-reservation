package reservation

import (
	"context"
	"errors"
	"time"
)

const (
	StatusActive    = "ACTIVE"
	StatusConfirmed = "CONFIRMED"
	StatusCanceled  = "CANCELED"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Create reservation (15 min hold)
func (s *Service) Create(
	ctx context.Context,
	productID int64,
	userID int64,
) (*Reservation, error) {

	expiresAt := time.Now().Add(15 * time.Minute)
	return s.repo.Create(ctx, productID, userID, expiresAt)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Reservation, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Confirm(ctx context.Context, id int64) error {
	return s.repo.UpdateStatus(ctx, id, StatusConfirmed)
}

func (s *Service) Cancel(ctx context.Context, id int64) error {
	return s.repo.UpdateStatus(ctx, id, StatusCanceled)
}

func (s *Service) List(
	ctx context.Context,
	userID *int64,
	status *string,
	limit int,
	offset int,
) ([]Reservation, error) {

	if limit <= 0 {
		return nil, errors.New("limit must be > 0")
	}

	return s.repo.List(ctx, userID, status, limit, offset)
}
