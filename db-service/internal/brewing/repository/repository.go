package repository

import (
	"alla/db-service/internal/transactor"
	"alla/db-service/models"
	errorList "alla/shared/errorList"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

//go:generate mockery --name=BrewingRepoInterface
type BrewingRepoInterface interface {
	CreateJob(ctx context.Context, m *models.BrewingJobs) error
	GetBrewStatus(ctx context.Context, uuid string) (string, error)
	SetStatus(ctx context.Context, uuid, status string, from []string) error
}

type BrewingRepo struct {
	db *sqlx.DB
}

func NewBrewingRepository(db *sqlx.DB) *BrewingRepo {
	return &BrewingRepo{
		db: db,
	}
}

func (r *BrewingRepo) CreateJob(ctx context.Context, m *models.BrewingJobs) error {
	var id string
	query := "INSERT INTO brewing_jobs (recipe_id, status, details) VALUES ($1, $2, $3) RETURNING public_id"

	err := r.db.QueryRowContext(ctx, query, m.RecipeID, m.Status, m.Details).Scan(&id)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == errorList.PgForeignKeyViolation {
			return errorList.ErrRecipeNotFound
		}
		return fmt.Errorf("CreateJob: %w", err)
	}

	m.PublicID = id
	return nil
}

func (r *BrewingRepo) GetBrewStatus(ctx context.Context, uuid string) (string, error) {
	query := `
		SELECT 
			status
		FROM 
			brewing_jobs
		WHERE 
			public_id = $1
	`
	var status string

	err := r.db.GetContext(ctx, &status, query, uuid)
	if err == sql.ErrNoRows {
		return "", errorList.ErrJobNotFound
	}
	if err != nil {
		return "", fmt.Errorf("GetBrewStatus: %w", err)
	}
	return status, nil
}

func (r *BrewingRepo) SetStatus(ctx context.Context, uuid, status string, from []string) error {

	tx := transactor.GetTx(ctx)

	query := `
		UPDATE 
			brewing_jobs
		SET
			status = $1::text,
			completed_at = CASE
				WHEN $1::text IN ('completed', 'failed') THEN now()
			END
		WHERE 
			public_id = $2 AND status = ANY($3)
	`
	var res sql.Result
	var err error
	if tx != nil {
		res, err = tx.ExecContext(ctx, query, status, uuid, pq.Array(from))
	} else {
		res, err = r.db.ExecContext(ctx, query, status, uuid, pq.Array(from))
	}
	if err != nil {
		return fmt.Errorf("SetStatus: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		if _, err := r.GetBrewStatus(ctx, uuid); err != nil {
			return err
		}
		return errorList.ErrInvalidStatusTransition
	}

	return nil
}
