package reservation

import (
	"context"
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates ACTIVE reservation
func (r *Repository) Create(
	ctx context.Context,
	productID int64,
	userID int64,
	expiresAt time.Time,
) (*Reservation, error) {

	query := `
		INSERT INTO reservations (product_id, user_id, status, expires_at)
		VALUES ($1, $2, 'ACTIVE', $3)
		RETURNING id, product_id, user_id, status, expires_at, created_at
	`

	var res Reservation
	err := r.db.QueryRowContext(
		ctx,
		query,
		productID,
		userID,
		expiresAt,
	).Scan(
		&res.ID,
		&res.ProductID,
		&res.UserID,
		&res.Status,
		&res.ExpiresAt,
		&res.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// GetByID returns reservation by id
func (r *Repository) GetByID(
	ctx context.Context,
	id int64,
) (*Reservation, error) {

	query := `
		SELECT id, product_id, user_id, status, expires_at, created_at
		FROM reservations
		WHERE id = $1
	`

	var res Reservation
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&res.ID,
		&res.ProductID,
		&res.UserID,
		&res.Status,
		&res.ExpiresAt,
		&res.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// UpdateStatus updates reservation status
func (r *Repository) UpdateStatusTx(
	ctx context.Context,
	tx *sql.Tx,
	id int64,
	status string,
) error {

	query := `
		UPDATE reservations
		SET status = $1
		WHERE id = $2
	`

	_, err := tx.ExecContext(ctx, query, status, id)
	return err
}

// List returns reservations with filters and pagination
func (r *Repository) List(
	ctx context.Context,
	userID *int64,
	status *string,
	limit int,
	offset int,
) ([]Reservation, error) {

	query := `
		SELECT id, product_id, user_id, status, expires_at, created_at
		FROM reservations
		WHERE ($1::bigint IS NULL OR user_id = $1)
		  AND ($2::text IS NULL OR status = $2)
		ORDER BY id DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		userID,
		status,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Reservation
	for rows.Next() {
		var res Reservation
		if err := rows.Scan(
			&res.ID,
			&res.ProductID,
			&res.UserID,
			&res.Status,
			&res.ExpiresAt,
			&res.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, res)
	}

	return result, nil
}

func (r *Repository) HasActiveReservationTx(
	ctx context.Context,
	tx *sql.Tx,
	productID, userID int64,
) (bool, error) {

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM reservations
			WHERE product_id = $1
			  AND user_id = $2
			  AND status = 'ACTIVE'
		)
	`

	var exists bool
	err := tx.QueryRowContext(ctx, query, productID, userID).Scan(&exists)
	return exists, err
}

func (r *Repository) CreateTx(
	ctx context.Context,
	tx *sql.Tx,
	productID, userID int64,
	expiresAt time.Time,
) (*Reservation, error) {

	query := `
		INSERT INTO reservations (product_id, user_id, status, expires_at)
		VALUES ($1, $2, 'ACTIVE', $3)
		RETURNING id, product_id, user_id, status, expires_at, created_at
	`

	var res Reservation
	err := tx.QueryRowContext(
		ctx,
		query,
		productID,
		userID,
		expiresAt,
	).Scan(
		&res.ID,
		&res.ProductID,
		&res.UserID,
		&res.Status,
		&res.ExpiresAt,
		&res.CreatedAt,
	)

	return &res, err
}

func (r *Repository) GetByIDForUpdate(
	ctx context.Context,
	tx *sql.Tx,
	id int64,
) (*Reservation, error) {

	query := `
		SELECT id, product_id, user_id, status, expires_at, created_at
		FROM reservations
		WHERE id = $1
		FOR UPDATE
	`

	var res Reservation
	err := tx.QueryRowContext(ctx, query, id).Scan(
		&res.ID,
		&res.ProductID,
		&res.UserID,
		&res.Status,
		&res.ExpiresAt,
		&res.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (r *Repository) GetExpiredForUpdate(
	ctx context.Context,
	tx *sql.Tx,
	now time.Time,
) ([]Reservation, error) {

	query := `
		SELECT id, product_id, user_id, status, expires_at, created_at
		FROM reservations
		WHERE status = 'ACTIVE'
		  AND expires_at < $1
		FOR UPDATE
	`

	rows, err := tx.QueryContext(ctx, query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Reservation
	for rows.Next() {
		var res Reservation
		if err := rows.Scan(
			&res.ID,
			&res.ProductID,
			&res.UserID,
			&res.Status,
			&res.ExpiresAt,
			&res.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, res)
	}

	return result, nil
}
