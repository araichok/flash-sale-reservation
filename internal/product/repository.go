package product

import (
	"context"
	"database/sql"
	"errors"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}
func (r *Repository) Create(
	ctx context.Context,
	name string,
	stock int,
) (*Product, error) {
	query := `
		INSERT INTO products (name, stock)
		VALUES ($1, $2)
		RETURNING id, name, stock, created_at
	`

	var p Product
	err := r.db.QueryRowContext(
		ctx,
		query,
		name,
		stock,
	).Scan(
		&p.ID,
		&p.Name,
		&p.Stock,
		&p.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *Repository) List(ctx context.Context) ([]Product, error) {
	query := `
		SELECT id, name, stock, created_at
		FROM products
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product

	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Stock,
			&p.CreatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

func (r *Repository) DecreaseStockTx(
	ctx context.Context,
	tx *sql.Tx,
	productID int64,
) error {

	query := `
		UPDATE products
		SET stock = stock - 1
		WHERE id = $1 AND stock > 0
	`

	res, err := tx.ExecContext(ctx, query, productID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("product out of stock")
	}

	return nil
}

func (r *Repository) IncreaseStockTx(
	ctx context.Context,
	tx *sql.Tx,
	productID int64,
) error {

	query := `
		UPDATE products
		SET stock = stock + 1
		WHERE id = $1
	`

	_, err := tx.ExecContext(ctx, query, productID)
	return err
}
