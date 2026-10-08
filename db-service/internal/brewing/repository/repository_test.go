package repository_test

import (
	alchemyRepo "alla/db-service/internal/alchemy/repository"
	brewingRepo "alla/db-service/internal/brewing/repository"
	"alla/db-service/models"
	"alla/shared/errorList"
	"alla/shared/status"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testBrewingRepo *brewingRepo.BrewingRepo
var testAlchemyRepo *alchemyRepo.AlchemyRepository
var testDB *sqlx.DB

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(m.Run())
	}

	ctx := context.Background()

	dbName := "testdb"
	dbUser := "Nya"
	dbPass := "bzzzz"

	pgContainer, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPass),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	testDB, err = sqlx.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	mig, err := migrate.New("file://../../db/migrations", connStr)
	if err != nil {
		log.Fatalf("failed create migrate: %v", err)
	}
	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed migrate up: %v", err)
	}

	testBrewingRepo = brewingRepo.NewBrewingRepository(testDB)
	testAlchemyRepo = alchemyRepo.NewAlchemyRepository(testDB)

	code := m.Run()

	pgContainer.Terminate(ctx)
	os.Exit(code)
}

func TestCreateJob(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		name           string
		wantErr        bool
		recipeID       int
		exprectedError error
	}{
		{
			name:     "Success",
			wantErr:  false,
			recipeID: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testDB.Exec("TRUNCATE ingredients, recipes, brewing_jobs   RESTART IDENTITY CASCADE")
			if err != nil {
				t.Fatalf("TestCreateJob: failed to truncate table: %v", err)
			}

			_, err = testAlchemyRepo.PostIngredients(context.Background(), models.Ingredient{
				Name:     "love",
				Quantity: 10,
			})
			assert.NoError(t, err)
			_, err = testAlchemyRepo.PostIngredients(context.Background(), models.Ingredient{
				Name:     "Soul",
				Quantity: 10,
			})
			assert.NoError(t, err)

			inputRecipe := models.Recipe{
				Name:               "love soul",
				BrewingTimeSeconds: 30,
				Ingredients: []models.RecipeIngredients{
					{
						IngredientID:   1,
						QuantityNeeded: 10,
					},
					{
						IngredientID:   2,
						QuantityNeeded: 4,
					},
				},
			}

			err = testAlchemyRepo.CreateRecipe(context.Background(), &inputRecipe)
			if err != nil {
				t.Fatalf("TestGetBrewStatus: CreateRecipe: %v", err)
			}

			var valueStrings = []string{}
			var valueArgs = []any{}
			for i, value := range inputRecipe.Ingredients {
				p1 := i*3 + 1
				p2 := i*3 + 2
				p3 := i*3 + 3
				valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", p1, p2, p3))
				valueArgs = append(valueArgs, inputRecipe.ID, value.IngredientID, value.QuantityNeeded)

			}

			err = testAlchemyRepo.CreateRecipeIngredients(context.Background(), valueStrings, valueArgs)
			if err != nil {
				t.Fatalf("TestGetBrewStatus: CreateRecipeIngredients: %v", err)
			}

			job := models.BrewingJobs{
				RecipeID: inputRecipe.ID,
				Status:   "queued",
			}

			err = testBrewingRepo.CreateJob(context.Background(), &job)

			if tt.wantErr {
				assert.ErrorIs(t, tt.exprectedError, err)
				return
			}
			assert.Equal(t, job.Status, "queued")
			assert.Equal(t, job.RecipeID, tt.recipeID)
			assert.NotEmpty(t, job.PublicID)
		})
	}
}

func TestSetStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		name            string
		initial         string
		to              string
		from            []string
		expectedErr     error
		wantStatus      string
		wantCompletedAt bool
		wantStartedAt   bool
	}{
		{
			name:          "queued -> brewing",
			initial:       status.StatusQueued,
			to:            status.StatusBrewing,
			from:          []string{status.StatusQueued},
			wantStatus:    status.StatusBrewing,
			wantStartedAt: true,
		},
		{
			name:            "brewing -> completed",
			initial:         status.StatusBrewing,
			to:              status.StatusCompleted,
			from:            []string{status.StatusBrewing},
			wantStatus:      status.StatusCompleted,
			wantCompletedAt: true,
		},
		{
			name:            "queued -> failed",
			initial:         status.StatusQueued,
			to:              status.StatusFailed,
			from:            []string{status.StatusQueued, status.StatusBrewing},
			wantStatus:      status.StatusFailed,
			wantCompletedAt: true,
		},
		{
			name:        "Error - completed -> failed",
			initial:     status.StatusCompleted,
			to:          status.StatusFailed,
			from:        []string{status.StatusQueued, status.StatusBrewing},
			expectedErr: errorList.ErrInvalidStatusTransition,
			wantStatus:  status.StatusCompleted,
		},
		{
			name:        "Error - queued -> completed",
			initial:     status.StatusQueued,
			to:          status.StatusCompleted,
			from:        []string{status.StatusBrewing},
			expectedErr: errorList.ErrInvalidStatusTransition,
			wantStatus:  status.StatusQueued,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uuid := createJob(t, tt.initial)

			err := testBrewingRepo.SetStatus(context.Background(), uuid, tt.to, tt.from)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			var job struct {
				Status      string       `db:"status"`
				CompletedAt sql.NullTime `db:"completed_at"`
				StartedAt   sql.NullTime `db:"started_at"`
			}
			err = testDB.Get(&job, "SELECT status, completed_at, started_at FROM brewing_jobs WHERE public_id = $1", uuid)
			if err != nil {
				t.Fatalf("read job: %v", err)
			}

			assert.Equal(t, tt.wantStatus, job.Status)
			assert.Equal(t, tt.wantCompletedAt, job.CompletedAt.Valid)
			assert.Equal(t, tt.wantStartedAt, job.StartedAt.Valid)
		})
	}

	t.Run("Error - job not found", func(t *testing.T) {
		createJob(t, status.StatusQueued)

		err := testBrewingRepo.SetStatus(context.Background(), "00000000-0000-0000-0000-000000000000",
			status.StatusBrewing, []string{status.StatusQueued})

		assert.ErrorIs(t, err, errorList.ErrJobNotFound)
	})
}

func createJob(t *testing.T, jobStatus string) string {
	t.Helper()

	_, err := testDB.Exec("TRUNCATE ingredients, recipes, recipe_ingredients, brewing_jobs RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}

	recipe := models.Recipe{
		Name:               "love soul",
		BrewingTimeSeconds: 30,
	}
	if err := testAlchemyRepo.CreateRecipe(context.Background(), &recipe); err != nil {
		t.Fatalf("create recipe: %v", err)
	}

	job := models.BrewingJobs{
		RecipeID: recipe.ID,
		Status:   jobStatus,
	}
	if err := testBrewingRepo.CreateJob(context.Background(), &job); err != nil {
		t.Fatalf("create job: %v", err)
	}
	return job.PublicID
}

func TestGetBrewStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		name        string
		wantErr     bool
		uuid        string
		expectedErr error
		status      string
	}{
		{
			name:    "Success",
			status:  "Nya",
			wantErr: false,
		},
		{
			name:        "Error - wrong uuid",
			uuid:        "00000000-0000-0000-0000-000000000000",
			status:      "",
			wantErr:     true,
			expectedErr: errorList.ErrJobNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testDB.Exec("TRUNCATE ingredients, recipes, recipe_ingredients, brewing_jobs RESTART IDENTITY CASCADE")
			if err != nil {
				t.Fatalf("Error truncate: %v", err)
			}

			_, err = testAlchemyRepo.PostIngredients(context.Background(), models.Ingredient{
				Name:     "love",
				Quantity: 10,
			})
			assert.NoError(t, err)
			_, err = testAlchemyRepo.PostIngredients(context.Background(), models.Ingredient{
				Name:     "Soul",
				Quantity: 10,
			})
			assert.NoError(t, err)

			inputRecipe := models.Recipe{
				Name:               "love soul",
				BrewingTimeSeconds: 30,
				Ingredients: []models.RecipeIngredients{
					{
						IngredientID:   1,
						QuantityNeeded: 10,
					},
					{
						IngredientID:   2,
						QuantityNeeded: 4,
					},
				},
			}

			err = testAlchemyRepo.CreateRecipe(context.Background(), &inputRecipe)
			if err != nil {
				t.Fatalf("TestGetBrewStatus: CreateRecipe: %v", err)
			}

			var valueStrings = []string{}
			var valueArgs = []any{}
			for i, value := range inputRecipe.Ingredients {
				p1 := i*3 + 1
				p2 := i*3 + 2
				p3 := i*3 + 3
				valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", p1, p2, p3))
				valueArgs = append(valueArgs, inputRecipe.ID, value.IngredientID, value.QuantityNeeded)

			}

			err = testAlchemyRepo.CreateRecipeIngredients(context.Background(), valueStrings, valueArgs)
			if err != nil {
				t.Fatalf("TestGetBrewStatus: CreateRecipeIngredients: %v", err)
			}

			job := models.BrewingJobs{
				RecipeID: inputRecipe.ID,
				Status:   tt.status,
			}
			err = testBrewingRepo.CreateJob(context.Background(), &job)

			if err != nil {
				t.Fatalf("TestGetBrewStatus: CreateJob: %v", err)
			}

			if tt.uuid != "" {
				job.PublicID = tt.uuid
			}

			status, err := testBrewingRepo.GetBrewStatus(context.Background(), job.PublicID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.status, status)
		})
	}
}

func TestGetRemaining(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	startBrewing := func(t *testing.T, uuid string) {
		err := testBrewingRepo.SetStatus(context.Background(), uuid, status.StatusBrewing, []string{status.StatusQueued})
		if err != nil {
			t.Fatalf("start brewing: %v", err)
		}
	}

	tests := []struct {
		name       string
		prepare    func(t *testing.T, uuid string)
		wantStatus string
		minRemain  int
		maxRemain  int
	}{
		{
			name:       "Just started",
			prepare:    startBrewing,
			wantStatus: status.StatusBrewing,
			minRemain:  29,
			maxRemain:  30,
		},
		{
			name: "Brewing time is over",
			prepare: func(t *testing.T, uuid string) {
				startBrewing(t, uuid)
				_, err := testDB.Exec("UPDATE brewing_jobs SET started_at = NOW() - interval '1 hour' WHERE public_id = $1", uuid)
				if err != nil {
					t.Fatalf("shift started_at: %v", err)
				}
			},
			wantStatus: status.StatusBrewing,
			minRemain:  0,
			maxRemain:  0,
		},
		{
			name:       "Not started",
			prepare:    func(t *testing.T, uuid string) {},
			wantStatus: status.StatusQueued,
			minRemain:  0,
			maxRemain:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uuid := createJob(t, status.StatusQueued)
			tt.prepare(t, uuid)

			gotStatus, remaining, err := testBrewingRepo.GetRemaining(context.Background(), uuid)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, gotStatus)
			assert.GreaterOrEqual(t, remaining, tt.minRemain)
			assert.LessOrEqual(t, remaining, tt.maxRemain)
		})
	}

	t.Run("Error - job not found", func(t *testing.T) {
		createJob(t, status.StatusQueued)

		_, _, err := testBrewingRepo.GetRemaining(context.Background(), "00000000-0000-0000-0000-000000000000")

		assert.ErrorIs(t, err, errorList.ErrJobNotFound)
	})
}
