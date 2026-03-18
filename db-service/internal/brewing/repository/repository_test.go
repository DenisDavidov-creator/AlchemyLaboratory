package repository_test

import (
	alchemyRepo "alla/db-service/internal/alchemy/repository"
	brewingRepo "alla/db-service/internal/brewing/repository"
	"alla/db-service/models"
	"alla/shared/errorList"
	"context"
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

	tests := []struct {
		name           string
		wantErr        bool
		recipeID       int
		exprectedError error
	}{
		{
			name:    "Success",
			wantErr: false,
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

			result, err := testBrewingRepo.CreateJob(context.Background(), &job)

			if tt.wantErr {
				assert.ErrorIs(t, tt.exprectedError, err)
				return
			}
			assert.Equal(t, result.Status, "queued")
			assert.Equal(t, result.ID, job.ID)
			assert.Equal(t, result.RecipeID, job.RecipeID)
		})
	}
}

func TestSetStatus(t *testing.T) {
	tests := []struct {
		name        string
		wantErr     bool
		uuid        string
		expectedErr error
		status      string
	}{
		{
			name:    "Success",
			wantErr: false,
			status:  "Nya",
		},
		{
			name:        "Error - wrong uuid",
			uuid:        "00000000-0000-0000-0000-000000000000",
			wantErr:     true,
			expectedErr: errorList.ErrJobNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testDB.Exec("TRUNCATE ingredients, recipes, recipe_ingredients RESTART IDENTITY CASCADE")
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
			result, err := testBrewingRepo.CreateJob(context.Background(), &job)

			if err != nil {
				t.Fatalf("TestGetBrewStatus: CreateJob: %v", err)
			}

			if tt.uuid != "" {
				result.PublicID = tt.uuid
			}

			err = testBrewingRepo.SetStatus(context.Background(), result.PublicID, tt.status)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, tt.expectedErr, err)
				}
				return
			}

			status, err := testBrewingRepo.GetBrewStatus(context.Background(), result.PublicID)
			if err != nil {
				t.Fatalf("Error - get status")
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.status, status)
		})
	}
}

func TestGetBrewStatus(t *testing.T) {
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
			result, err := testBrewingRepo.CreateJob(context.Background(), &job)

			if err != nil {
				t.Fatalf("TestGetBrewStatus: CreateJob: %v", err)
			}

			if tt.uuid != "" {
				result.PublicID = tt.uuid
			}

			status, err := testBrewingRepo.GetBrewStatus(context.Background(), result.PublicID)

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
