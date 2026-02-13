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

// 1. Create reservation (ACTIVE)
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

	return &res, err
}

// 2. Get by ID
func (r *Repository) GetByID(ctx context.Context, id int64) (*Reservation, error) {
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

	return &res, err
}

// 3. Update status
func (r *Repository) UpdateStatus(
	ctx context.Context,
	id int64,
	status string,
) error {
	query := `
	UPDATE reservations
	SET status = $1
	WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

// 4. List with filters + pagination
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
		var r Reservation
		if err := rows.Scan(
			&r.ID,
			&r.ProductID,
			&r.UserID,
			&r.Status,
			&r.ExpiresAt,
			&r.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, r)
	}

	return result, nil
}
