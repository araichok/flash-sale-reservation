package reservation

import (
	"context"
	"errors"
	"flash-sale-reservation/internal/product"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

const (
	StatusActive    = "ACTIVE"
	StatusConfirmed = "CONFIRMED"
	StatusCanceled  = "CANCELED"
)

type Service struct {
	repo        *Repository
	productRepo *product.Repository
	redis       *redis.Client
}

func NewService(
	repo *Repository,
	productRepo *product.Repository,
	redis *redis.Client,
) *Service {
	return &Service{
		repo:        repo,
		productRepo: productRepo,
		redis:       redis,
	}
}

// Create reservation (15 min hold)
func (s *Service) Create(
	ctx context.Context,
	productID int64,
	userID int64,
) (*Reservation, error) {

	tx, err := s.repo.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Проверка активного резерва
	hasActive, err := s.repo.HasActiveReservationTx(ctx, tx, productID, userID)
	if err != nil {
		return nil, err
	}
	if hasActive {
		return nil, errors.New("active reservation already exists")
	}

	// 2. Уменьшаем stock продукта
	if err := s.productRepo.DecreaseStockTx(ctx, tx, productID); err != nil {
		return nil, err
	}

	// 3. Создаём резерв на 5 минут
	expiresAt := time.Now().Add(5 * time.Minute)

	res, err := s.repo.CreateTx(ctx, tx, productID, userID, expiresAt)
	if err != nil {
		return nil, err
	}

	// 4. Commit
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 5. Redis TTL
	key := fmt.Sprintf("reservation:%d", res.ID)
	ttl := time.Until(res.ExpiresAt)
	_ = s.redis.Set(ctx, key, "active", ttl).Err()

	// 6. Redis metric
	_ = s.redis.Incr(ctx, "metrics:reservations:created").Err()

	return res, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Reservation, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Confirm(ctx context.Context, id int64) error {

	tx, err := s.repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := s.repo.GetByIDForUpdate(ctx, tx, id)
	if err != nil {
		return err
	}

	if res.Status != StatusActive {
		return errors.New("only ACTIVE reservation can be confirmed")
	}

	if err := s.repo.UpdateStatus(ctx, id, StatusConfirmed); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// Redis metric
	_ = s.redis.Incr(ctx, "metrics:reservations:confirmed").Err()

	return nil
}

func (s *Service) Cancel(ctx context.Context, id int64) error {

	tx, err := s.repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := s.repo.GetByIDForUpdate(ctx, tx, id)
	if err != nil {
		return err
	}

	if res.Status != StatusActive {
		return errors.New("only ACTIVE reservation can be canceled")
	}

	// Возвращаем stock
	if err := s.productRepo.IncreaseStockTx(ctx, tx, res.ProductID); err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, id, StatusCanceled); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// Redis metric
	_ = s.redis.Incr(ctx, "metrics:reservations:canceled").Err()

	return nil
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
