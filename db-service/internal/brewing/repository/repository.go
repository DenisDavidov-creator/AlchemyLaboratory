package repository

import (
	"alla/db-service/internal/transactor"
	"alla/db-service/models"
	errorList "alla/shared/errorList"
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

//go:generate mockery --name=BrewingRepoInterface
type BrewingRepoInterface interface {
	CreateJob(ctx context.Context, m *models.BrewingJobs) (*models.BrewingJobs, error)
	GetBrewStatus(ctx context.Context, uuid string) (string, error)
	SetStatus(ctx context.Context, uuid, status string) error
}

type BrewingRepo struct {
	db *sqlx.DB
}

func NewBrewingRepository(db *sqlx.DB) *BrewingRepo {
	return &BrewingRepo{
		db: db,
	}
}

func (r *BrewingRepo) CreateJob(ctx context.Context, m *models.BrewingJobs) (*models.BrewingJobs, error) {
	var id string
	query := "INSERT INTO brewing_jobs (recipe_id, status, details) VALUES ($1, $2, $3) RETURNING public_id"

	err := r.db.QueryRowContext(ctx, query, m.RecipeID, m.Status, m.Details).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("CreateJob: %w", err)
	}

	m.PublicID = id
	return m, nil
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

func (r *BrewingRepo) SetStatus(ctx context.Context, uuid, status string) error {

	tx := transactor.GetTx(ctx)

	query := `
		UPDATE 
			brewing_jobs
		SET
			status = $1,
			completed_at = NOW()
		WHERE 
			public_id = $2
	`
	var res sql.Result
	var err error
	if tx != nil {
		res, err = tx.Exec(query, status, uuid)
	} else {
		res, err = r.db.Exec(query, status, uuid)
	}
	if err != nil {
		return fmt.Errorf("SetStatus: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errorList.ErrJobNotFound
	}

	return nil
}
