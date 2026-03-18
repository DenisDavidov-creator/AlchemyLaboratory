package repository_test

import (
	"alla/db-service/internal/alchemy/repository"
	"alla/db-service/internal/transactor"
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
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testRepo repository.AlchemyRepository
var testTrans transactor.PTransactor
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

	testRepo = *repository.NewAlchemyRepository(testDB)
	testTrans = *transactor.NewPtransactor(testDB)
	code := m.Run()

	pgContainer.Terminate(ctx)
	os.Exit(code)
}

func TestCheckIngredientExistsByName(t *testing.T) {
	tests := []struct {
		name           string
		ingName        string
		expectedResult bool
		wantErr        bool
		expectedErr    error
		setup          func()
	}{
		{name: "Success - exist ingredient",
			ingName: "Glass",
			wantErr: false,
			setup: func() {
				_, err := testRepo.PostIngredients(context.Background(), models.Ingredient{
					Name:     "Glass",
					Quantity: 10,
				})
				require.NoError(t, err)

			},
			expectedResult: true,
		},
		{name: "Success - doesn't exist ingredient",
			ingName: "Stone",
			wantErr: false,
			setup: func() {
				_, err := testRepo.PostIngredients(context.Background(), models.Ingredient{
					Name:     "Glass",
					Quantity: 10,
				})
				require.NoError(t, err)

			},
			expectedResult: false,
		},
		{name: "Success - an empty name",
			ingName: "",
			wantErr: false,
			setup: func() {
				_, err := testRepo.PostIngredients(context.Background(), models.Ingredient{
					Name:     "Glass",
					Quantity: 10,
				})
				require.NoError(t, err)

			},
			expectedResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testDB.Exec("TRUNCATE ingredients RESTART IDENTITY CASCADE")
			if err != nil {
				t.Fatalf("failed to truncate table: %v", err)
			}

			if tt.setup != nil {
				tt.setup()
			}
			ctx := context.Background()

			result, err := testRepo.CheckIngredientExistsByName(ctx, tt.ingName)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)

		})
	}
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

			_, err := testDB.Exec("TRUNCATE ingredients RESTART IDENTITY CASCADE")
			if err != nil {
				t.Fatalf("failed to truncate table: %v", err)
			}

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
			expectedErr:     errorList.ErrIngredientNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err := testDB.Exec("TRUNCATE ingredients RESTART IDENTITY CASCADE")
			if err != nil {
				t.Fatalf("failed to truncate table: %v", err)
			}

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
				t.Errorf("GetIngredients() error = %v, wantErr %v", err, tt.wantErr)
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

func TestGetRecipes(t *testing.T) {

	tests := []struct {
		name          string
		inputData     models.Recipe
		wantErr       bool
		expectedError error
		expectedData  models.Recipe
	}{
		{
			name: "Success",
			inputData: models.Recipe{
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
			},
			wantErr: false,
			expectedData: models.Recipe{
				ID:                 1,
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
			},
		},
		{
			name:         "Success - empty slice",
			inputData:    models.Recipe{},
			wantErr:      false,
			expectedData: models.Recipe{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testDB.Exec("TRUNCATE ingredients, recipes, brewing_jobs   RESTART IDENTITY CASCADE")
			if err != nil {
				t.Fatalf("failed to truncate table: %v", err)
			}
			if tt.inputData.Name != "" {

				_, err = testRepo.PostIngredients(context.Background(), models.Ingredient{
					Name:     "love",
					Quantity: 10,
				})
				assert.NoError(t, err)
				_, err = testRepo.PostIngredients(context.Background(), models.Ingredient{
					Name:     "Soul",
					Quantity: 10,
				})
				assert.NoError(t, err)

				inputRecipe := tt.inputData

				err = testRepo.CreateRecipe(context.Background(), &inputRecipe)
				if err != nil {
					t.Fatalf("GetRecipe: CreateRecipe: %v", err)
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

				err = testRepo.CreateRecipeIngredients(context.Background(), valueStrings, valueArgs)
				if err != nil {
					t.Fatalf("GetRecipe: CreateRecipeIngredients: %v", err)
				}
			}

			result, err := testRepo.GetRecipes(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.ErrorIs(t, err, tt.expectedError)
				}
				return
			}

			assert.NoError(t, err)
			if tt.inputData.Name == "" {
				assert.Empty(t, result)
				return
			}
			for _, value := range result {
				assert.Equal(t, value.ID, tt.expectedData.ID)
				assert.Equal(t, value.Name, tt.expectedData.Name)
				assert.Equal(t, value.Description, tt.expectedData.Description)
				assert.Equal(t, value.BrewingTimeSeconds, tt.expectedData.BrewingTimeSeconds)
				for i, ingValue := range value.Ingredients {
					assert.Equal(t, ingValue.IngredientID, tt.expectedData.Ingredients[i].IngredientID)
					assert.Equal(t, ingValue.QuantityNeeded, tt.expectedData.Ingredients[i].QuantityNeeded)
				}

			}
		})
	}
}

func TestCreateRecipe(t *testing.T) {
	tests := []struct {
		name        string
		recipeName  string
		timeSecond  int
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "Success",
			recipeName: "Love soul",
			timeSecond: 30,
			wantErr:    false,
		},
		{
			name:       "Rollback - recipe not saved on tx roolback",
			recipeName: "Love soul",
			timeSecond: 10,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testDB.Exec("TRUNCATE recipes RESTART IDENTITY CASCADE")
			if err != nil {
				t.Fatalf("Error truncate: %v", err)
			}
			createingRicype := models.Recipe{
				Name:               tt.recipeName,
				BrewingTimeSeconds: tt.timeSecond,
			}

			if tt.name == "Rollback - recipe not saved on tx roolback" {
				tx, err := testDB.Beginx()
				if err != nil {
					t.Fatalf("begin tx: %v", err)
				}

				ctx := transactor.InjectTx(context.Background(), tx)
				err = testRepo.CreateRecipe(ctx, &createingRicype)
				assert.NoError(t, err)

				tx.Rollback()

				var bridgeCount int
				err = testDB.Get(&bridgeCount, "SELECT count(*) FROM recipes")
				assert.NoError(t, err)

				assert.Equal(t, 0, bridgeCount, "Should have 1 irecipe")
			}

			err = testRepo.CreateRecipe(context.Background(), &createingRicype)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				return
			}

			assert.NoError(t, err)

			var bridgeCount int
			err = testDB.Get(&bridgeCount, "SELECT count(*) FROM recipes")
			assert.NoError(t, err)

			assert.Equal(t, 1, bridgeCount, "Should have 1 irecipe")

		})
	}
}

func TestCreateRecipeIngredients(t *testing.T) {
	tests := []struct {
		name        string
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "Success",
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name:        "Rollback - recipe-ingredient not saved on tx roolback",
			wantErr:     false,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testDB.Exec("TRUNCATE recipes, ingredients, recipe_ingredients RESTART IDENTITY CASCADE")
			if err != nil {
				t.Fatalf("Error truncate: %v", err)
			}
			ing1 := models.Ingredient{
				Name:     "stone",
				Quantity: 10,
			}
			ing2 := models.Ingredient{
				Name:     "Key",
				Quantity: 2,
			}

			_, err = testRepo.PostIngredients(context.Background(), ing1)
			assert.NoError(t, err)
			_, err = testRepo.PostIngredients(context.Background(), ing2)
			assert.NoError(t, err)

			createingRicype := models.Recipe{
				Name:               "Soul",
				BrewingTimeSeconds: 10,
			}

			err = testRepo.CreateRecipe(context.Background(), &createingRicype)
			assert.NoError(t, err)

			ValueString := []string{"($1, $2, $3)", "($4, $5, $6)"}
			ValueArgs := []any{1, 1, 10, 1, 2, 4}

			if tt.name == "Rollback - recipe-ingredient not saved on tx roolback" {
				tx, err := testDB.Beginx()
				if err != nil {
					t.Fatalf("begin tx: %v", err)
				}
				ctx := transactor.InjectTx(context.Background(), tx)

				err = testRepo.CreateRecipeIngredients(ctx, ValueString, ValueArgs)
				assert.NoError(t, err)

				tx.Rollback()

				var bridgeCount int
				err = testDB.Get(&bridgeCount, "SELECT count(*) FROM recipe_ingredients")
				assert.NoError(t, err)

				assert.Equal(t, 0, bridgeCount, "Should have 0 irecipe")
			}

			err = testRepo.CreateRecipeIngredients(context.Background(), ValueString, ValueArgs)
			assert.NoError(t, err)

			recipesList, err := testRepo.GetRecipes(context.Background())
			assert.NoError(t, err)

			assert.Equal(t, createingRicype.BrewingTimeSeconds, recipesList[0].BrewingTimeSeconds)
			assert.Equal(t, 1, recipesList[0].ID)
			assert.Equal(t, createingRicype.Name, recipesList[0].Name)
		})
	}
}

func TestGetRicpeID(t *testing.T) {
	tests := []struct {
		name        string
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "Success",
			wantErr:     false,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testDB.Exec("TRUNCATE recipes, ingredients, recipe_ingredients RESTART IDENTITY CASCADE")
			if err != nil {
				t.Fatalf("Error truncate: %v", err)
			}
			ing1 := models.Ingredient{
				Name:     "stone",
				Quantity: 10,
			}
			ing2 := models.Ingredient{
				Name:     "Key",
				Quantity: 2,
			}

			_, err = testRepo.PostIngredients(context.Background(), ing1)
			assert.NoError(t, err)
			_, err = testRepo.PostIngredients(context.Background(), ing2)
			assert.NoError(t, err)

			createingRicype := models.Recipe{
				Name:               "Soul",
				BrewingTimeSeconds: 10,
			}

			err = testRepo.CreateRecipe(context.Background(), &createingRicype)
			assert.NoError(t, err)

			ValueString := []string{"($1, $2, $3)", "($4, $5, $6)"}
			ValueArgs := []any{1, 1, 10, 1, 2, 4}

			if tt.name == "Rollback - recipe-ingredient not saved on tx roolback" {
				tx, err := testDB.Beginx()
				if err != nil {
					t.Fatalf("begin tx: %v", err)
				}
				ctx := transactor.InjectTx(context.Background(), tx)

				err = testRepo.CreateRecipeIngredients(ctx, ValueString, ValueArgs)
				assert.NoError(t, err)

				tx.Rollback()

				var bridgeCount int
				err = testDB.Get(&bridgeCount, "SELECT count(*) FROM recipe_ingredients")
				assert.NoError(t, err)

				assert.Equal(t, 0, bridgeCount, "Should have 0 irecipe")
			}

			err = testRepo.CreateRecipeIngredients(context.Background(), ValueString, ValueArgs)
			assert.NoError(t, err)

			recipesList, err := testRepo.GetRecipes(context.Background())
			assert.NoError(t, err)

			assert.Equal(t, createingRicype.BrewingTimeSeconds, recipesList[0].BrewingTimeSeconds)
			assert.Equal(t, 1, recipesList[0].ID)
			assert.Equal(t, createingRicype.Name, recipesList[0].Name)
		})
	}
}

func TestCheckingIngredients(t *testing.T) {
	tests := []struct {
		name        string
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "Success",
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name:        "Rollback - ingredients not increase",
			wantErr:     false,
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := testDB.Exec("TRUNCATE ingredients RESTART IDENTITY CASCADE")
			if err != nil {
				t.Fatalf("Error truncate: %v", err)
			}
			seed1 := models.Ingredient{
				Name:     "stone",
				Quantity: 10,
			}
			seed2 := models.Ingredient{
				Name:     "Key",
				Quantity: 2,
			}

			ing1, err := testRepo.PostIngredients(context.Background(), seed1)
			assert.NoError(t, err)
			ing2, err := testRepo.PostIngredients(context.Background(), seed2)
			assert.NoError(t, err)

			req := []models.RecipeIngredients{
				{IngredientID: ing1.ID, QuantityNeeded: 1},
				{IngredientID: ing2.ID, QuantityNeeded: 1},
			}
			if tt.name == "Rollback - ingredients not increase" {
				tx, err := testDB.Beginx()
				if err != nil {
					t.Fatalf("begin tx: %v", err)
				}
				ctx := transactor.InjectTx(context.Background(), tx)

				err = testRepo.CheckingIngridients(ctx, req)
				assert.NoError(t, err)
				result, err := testRepo.GetIngredients(ctx)
				assert.NoError(t, err)

				tx.Rollback()

				assert.Equal(t, seed1.Quantity, result[0].Quantity)
				assert.Equal(t, seed2.Quantity, result[1].Quantity)

				return
			}

			err = testRepo.CheckingIngridients(context.Background(), req)
			assert.NoError(t, err)

			result, err := testRepo.GetIngredients(context.Background())
			assert.NoError(t, err)

			assert.Equal(t, seed1.Quantity-1, result[0].Quantity)
			assert.Equal(t, seed2.Quantity-1, result[1].Quantity)

		})
	}
}
