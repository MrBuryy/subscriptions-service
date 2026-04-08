package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/MrBuryy/subscriptions-service/internal/model"
	"github.com/MrBuryy/subscriptions-service/internal/repository"
	"github.com/google/uuid"
)

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{
		db: db,
	}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *model.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			service_name,
			price,
			user_id,
			start_date,
			end_date
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at;
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
	).Scan(
		&sub.ID,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	query := `
		SELECT
			id,
			service_name,
			price,
			user_id,
			start_date,
			end_date,
			created_at,
			updated_at
		FROM subscriptions
		WHERE id = $1;
	`

	var sub model.Subscription
	var endDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&endDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	if endDate.Valid {
		sub.EndDate = &endDate.Time
	}

	return &sub, nil
}

func (r *SubscriptionRepository) List(ctx context.Context, filter model.SubscriptionFilter) ([]model.Subscription, error) {
	var sb strings.Builder

	sb.WriteString(`
		SELECT
			id,
			service_name,
			price,
			user_id,
			start_date,
			end_date,
			created_at,
			updated_at
		FROM subscriptions
		WHERE 1=1
	`)

	args := make([]any, 0, 4)
	argPos := 1

	if filter.UserID != nil {
		sb.WriteString(fmt.Sprintf(" AND user_id = $%d", argPos))
		args = append(args, *filter.UserID)
		argPos++
	}

	if filter.ServiceName != nil {
		sb.WriteString(fmt.Sprintf(" AND service_name = $%d", argPos))
		args = append(args, *filter.ServiceName)
		argPos++
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}

	sb.WriteString(fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1))
	args = append(args, limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subs := make([]model.Subscription, 0)

	for rows.Next() {
		var sub model.Subscription
		var endDate sql.NullTime

		if err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&endDate,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if endDate.Valid {
			t := endDate.Time
			sub.EndDate = &t
		}

		subs = append(subs, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return subs, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub *model.Subscription) error {
	query := `
		UPDATE subscriptions
		SET
			service_name = $1,
			price = $2,
			user_id = $3,
			start_date = $4,
			end_date = $5,
			updated_at = now()
		WHERE id = $6
		RETURNING updated_at;
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
		sub.ID,
	).Scan(&sub.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.ErrNotFound
		}
		return err
	}

	return nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM subscriptions
		WHERE id = $1;
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}
