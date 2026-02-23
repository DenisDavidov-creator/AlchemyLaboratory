package repository_test

import (
	"alchemicallabaratory/errorList"
	"alchemicallabaratory/models"
	"alchemicallabaratory/repository"
	"context"
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

var testRepo *repository.GrimoireRepository
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

	mig, err := migrate.New("file://../db/migrations", connStr)
	if err != nil {
		log.Fatalf("failed create migrate: %v", err)
	}
	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed migrate up: %v", err)
	}

	testRepo = repository.NewGrimoireRepository(testDB)

	code := m.Run()

	pgContainer.Terminate(ctx)
	os.Exit(code)
}

func TestPostIngredeints(t *testing.T) {

	tests := []struct {
		name     string
		ingModel models.Ingredient
		wantErr  bool
		setup    func()
	}{
		{
			name: "Success",
			ingModel: models.Ingredient{
				Name:        "grass",
				Description: "",
				Quantity:    10,
			},
			wantErr: false,
		},
		{
			name: "Error - the same name",
			ingModel: models.Ingredient{
				Name:        "mandrake",
				Description: "Screams a lot",
				Quantity:    10,
			},
			setup: func() {
				testRepo.PostIngredients(context.Background(), models.Ingredient{Name: "mandrake", Quantity: 5})
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testDB.Exec("DELETE FROM ingredients")

			if tt.setup != nil {
				tt.setup()
			}

			result, err := testRepo.PostIngredients(context.Background(), tt.ingModel)

			if (err != nil) != tt.wantErr {
				t.Errorf("PostIngredients() error = %v, wantErr %v", err, tt.wantErr)

			}

			if !tt.wantErr && result.ID == 0 {
				t.Error("Expected DB to return a real ID, but got 0")
			}
			if tt.wantErr == false {
				assert.Equal(t, tt.ingModel.Name, result.Name)
				assert.Equal(t, 1, result.ID)
				assert.Equal(t, tt.ingModel.Quantity, result.Quantity)
			}
		})
	}
}

func TestAddIngredeints(t *testing.T) {

	tests := []struct {
		name             string
		quantity         int
		overrideID       int
		initialQuantity  int
		expectedQuantity int
		wantErr          bool
		expectedErr      error
	}{
		{
			name:             "Success",
			initialQuantity:  5,
			quantity:         10,
			expectedQuantity: 15,
			wantErr:          false,
		},
		{
			name:            "Error - non-existing ingredient",
			overrideID:      100000,
			initialQuantity: 1,
			quantity:        10,
			wantErr:         true,
			expectedErr:     errorList.ErrAddIngredientsNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testDB.Exec("TRUNCATE ingredients RESTART IDENTITY")
			testDB.Exec("DELETE FROM ingredients")
			ing, err := testRepo.PostIngredients(context.Background(), models.Ingredient{
				Name:        "Dragon eye",
				Description: "",
				Quantity:    tt.initialQuantity,
			})
			if err != nil {
				t.Fatalf("PostIngredient err: %v", err)
			}
			assert.NoError(t, err)

			targetID := ing.ID
			if tt.overrideID != 0 {
				targetID = tt.overrideID
			}

			err = testRepo.AddIngredients(context.Background(), targetID, tt.quantity)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, tt.expectedErr, err)

				}

				return
			}
			assert.NoError(t, err)
			m, _ := testRepo.GetIngredients(context.Background())
			assert.Equal(t, tt.expectedQuantity, m[0].Quantity)

		})
	}
}

func TestGetIngredeints(t *testing.T) {

	tests := []struct {
		name     string
		seedData []models.Ingredient
		wantErr  bool
	}{
		{
			name: "Success",
			seedData: []models.Ingredient{
				{
					Name:     "Dragon's eye",
					Quantity: 4,
				},
				{
					Name:     "Dragon's poison",
					Quantity: 4,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testDB.Exec("TRUNCATE ingredients RESTART IDENTITY CASCADE")
			if err != nil {
				t.Fatalf("failed to truncate table: %v", err)
			}

			var createdIngredients []models.Ingredient

			for _, seed := range tt.seedData {
				ing, err := testRepo.PostIngredients(context.Background(), seed)
				if err != nil {
					t.Fatalf("Failed seed %v", err)
				}
				createdIngredients = append(createdIngredients, *ing)
			}

			result, err := testRepo.GetIngredients(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("AddIngredients() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Equal(t, len(createdIngredients), len(result))

				for i := range result {
					assert.Equal(t, createdIngredients[i].ID, result[i].ID)
					assert.Equal(t, createdIngredients[i].Name, result[i].Name)
					assert.Equal(t, createdIngredients[i].Quantity, result[i].Quantity)
				}

			}

		})
	}
}

func TestPostJob(t *testing.T) {

	tests := []struct {
		name           string
		inputData      models.BrewingJobs
		wantErr        bool
		recipeID       int
		exprectedError error
	}{
		{
			name: "Success",
			inputData: models.BrewingJobs{
				Status: "queued",
			},
			wantErr: false,
		},
		{
			name: "Error don't exist RecipeID",
			inputData: models.BrewingJobs{
				Status: "queued",
			},
			wantErr:        true,
			recipeID:       10000,
			exprectedError: errorList.ErrPostJob,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testDB.Exec("TRUNCATE ingredients, recipes, brewing_jobs   RESTART IDENTITY CASCADE")
			if err != nil {
				t.Fatalf("failed to truncate table: %v", err)
			}

			ing1, err := testRepo.PostIngredients(context.Background(), models.Ingredient{
				Name:     "love",
				Quantity: 10,
			})
			assert.NoError(t, err)
			ing2, err := testRepo.PostIngredients(context.Background(), models.Ingredient{
				Name:     "Soul",
				Quantity: 10,
			})
			assert.NoError(t, err)

			recipe, err := testRepo.PostRecipe(context.Background(), models.Recipe{
				Name:               "love soul",
				BrewingTimeSeconds: 10,
				Ingredients: []models.RecipeIngredients{
					{
						IngredientID:   ing1.ID,
						QuantityNeeded: 2,
					},
					{
						IngredientID:   ing2.ID,
						QuantityNeeded: 2,
					},
				},
			})

			if err != nil {
				t.Fatalf("Error - post recipe")
			}
			jobInput := tt.inputData
			if tt.recipeID != 0 {
				jobInput.RecipeID = tt.recipeID
			} else {
				jobInput.RecipeID = recipe.ID
			}
			result, err := testRepo.PostJob(context.Background(), jobInput)

			if tt.wantErr {
				assert.ErrorIs(t, tt.exprectedError, err)
				return
			}
			assert.Equal(t, result.Status, "queued")
			assert.Equal(t, result.ID, jobInput.ID)
			assert.Equal(t, result.RecipeID, jobInput.RecipeID)
		})
	}
}

func TestPostRecipe(t *testing.T) {
	tests := []struct {
		name                 string
		recipeName           string
		timeSecond           int
		overriedIngredientID int
		wantErrCreateRecipe  bool
		wantErr              bool
		expectedErr          error
	}{
		{
			name:       "Success",
			recipeName: "Love soul",
			timeSecond: 30,
			wantErr:    false,
		},
		{
			name:                 "Error - wrong ingredient ID",
			recipeName:           "Love soul",
			overriedIngredientID: 99999,
			timeSecond:           30,
			wantErr:              true,
			expectedErr:          errorList.ErrCreateConnectionRecipeIngredient,
		},
		{
			name:                "Error - wrong timeSeconds",
			recipeName:          "Love soul",
			wantErrCreateRecipe: true,
			timeSecond:          30,
			wantErr:             true,
			expectedErr:         errorList.ErrCreateRecipe,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testDB.Exec("TRUNCATE ingredients, recipes, recipe_ingredients RESTART IDENTITY CASCADE")
			if err != nil {
				t.Fatalf("Error truncate: %v", err)
			}

			ing1, err := testRepo.PostIngredients(context.Background(), models.Ingredient{
				Name:     "love",
				Quantity: 10,
			})
			assert.NoError(t, err)
			ing2, err := testRepo.PostIngredients(context.Background(), models.Ingredient{
				Name:     "Soul",
				Quantity: 10,
			})
			assert.NoError(t, err)

			if tt.overriedIngredientID != 0 {
				ing1.ID = tt.overriedIngredientID
			}

			createingRicype := models.Recipe{
				Name:               tt.recipeName,
				BrewingTimeSeconds: tt.timeSecond,
				Ingredients: []models.RecipeIngredients{
					{IngredientID: ing1.ID, QuantityNeeded: 3},
					{IngredientID: ing2.ID, QuantityNeeded: 3},
				},
			}
			if tt.wantErrCreateRecipe {
				createingRicype.BrewingTimeSeconds = -1
			}
			result, err := testRepo.PostRecipe(context.Background(), createingRicype)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				return
			}

			assert.NoError(t, err)

			var bridgeCount int
			err = testDB.Get(&bridgeCount, "SELECT count(*) FROM recipe_ingredients WHERE recipe_id = $1", result.ID)
			assert.Equal(t, 2, bridgeCount, "Should have 2 ingredient")

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
			uuid:        "I-kissed-a-girl",
			wantErr:     true,
			expectedErr: errorList.ErrSetStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testDB.Exec("TRUNCATE ingredients, recipes, recipe_ingredients RESTART IDENTITY CASCADE")
			if err != nil {
				t.Fatalf("Error truncate: %v", err)
			}

			ing1, err := testRepo.PostIngredients(context.Background(), models.Ingredient{
				Name:     "love",
				Quantity: 10,
			})
			assert.NoError(t, err)
			ing2, err := testRepo.PostIngredients(context.Background(), models.Ingredient{
				Name:     "Soul",
				Quantity: 10,
			})
			assert.NoError(t, err)

			recipe, err := testRepo.PostRecipe(context.Background(), models.Recipe{
				Name:               "love soul",
				BrewingTimeSeconds: 10,
				Ingredients: []models.RecipeIngredients{
					{
						IngredientID:   ing1.ID,
						QuantityNeeded: 2,
					},
					{
						IngredientID:   ing2.ID,
						QuantityNeeded: 2,
					},
				},
			})

			if err != nil {
				t.Fatalf("Error - post recipe")
			}
			job := models.BrewingJobs{
				RecipeID: recipe.ID,
				Status:   "quiued",
			}
			result, err := testRepo.PostJob(context.Background(), job)
			if err != nil {
				t.Fatalf("Error - post jobs")
			}

			if tt.uuid != "" {
				result.PublicID = tt.uuid
			}

			err = testRepo.SetStatus(context.Background(), result.PublicID, tt.status)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, tt.expectedErr, err)
				}
				return
			}

			status, err := testRepo.GetBrewStatus(context.Background(), result.PublicID)
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
			uuid:        "I-kissed-a-girl",
			status:      "",
			wantErr:     true,
			expectedErr: errorList.ErrGetStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testDB.Exec("TRUNCATE ingredients, recipes, recipe_ingredients RESTART IDENTITY CASCADE")
			if err != nil {
				t.Fatalf("Error truncate: %v", err)
			}

			ing1, err := testRepo.PostIngredients(context.Background(), models.Ingredient{
				Name:     "love",
				Quantity: 10,
			})
			assert.NoError(t, err)
			ing2, err := testRepo.PostIngredients(context.Background(), models.Ingredient{
				Name:     "Soul",
				Quantity: 10,
			})
			assert.NoError(t, err)

			recipe, err := testRepo.PostRecipe(context.Background(), models.Recipe{
				Name:               "love soul",
				BrewingTimeSeconds: 10,
				Ingredients: []models.RecipeIngredients{
					{
						IngredientID:   ing1.ID,
						QuantityNeeded: 2,
					},
					{
						IngredientID:   ing2.ID,
						QuantityNeeded: 2,
					},
				},
			})

			if err != nil {
				t.Fatalf("Error - post recipe")
			}
			job := models.BrewingJobs{
				RecipeID: recipe.ID,
				Status:   tt.status,
			}
			result, err := testRepo.PostJob(context.Background(), job)
			if err != nil {
				t.Fatalf("Error - post jobs")
			}

			if tt.uuid != "" {
				result.PublicID = tt.uuid
			}

			status, err := testRepo.GetBrewStatus(context.Background(), result.PublicID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, tt.expectedErr, err)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, status, tt.status)
		})
	}
}
