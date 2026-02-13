package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) InsertTx(
	ctx context.Context,
	tx *sql.Tx,
	eventType string,
	payload any,
) error {

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO outbox_events (event_type, payload)
		VALUES ($1, $2)
	`

	_, err = tx.ExecContext(ctx, query, eventType, data)
	return err
}
