package service_test

import (
	repoMock "alla/db-service/internal/alchemy/repository/mocks"
	"alla/db-service/internal/alchemy/service"
	transMock "alla/db-service/internal/transactor/mocks"
	"alla/db-service/models"
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var DB_ERROR = errors.New("DB_ERROR")

func TestPostIngredients(t *testing.T) {
	tests := []struct {
		Name           string
		IngDTO         dto.IngredientDTO
		IngModel       models.Ingredient
		CheckErr       error
		CheckResult    bool
		ExpectedResult models.Ingredient
		ExpectedErr    error
	}{
		{
			Name: "Success",
			IngDTO: dto.IngredientDTO{
				Name:        "Soul",
				Description: "",
				Quantity:    10,
			},
			IngModel: models.Ingredient{
				Name:        "Soul",
				Description: "",
				Quantity:    10,
			},
			ExpectedResult: models.Ingredient{
				ID:          1,
				Name:        "Soul",
				Description: "",
				Quantity:    10,
			},
			ExpectedErr: nil,
		},
		{
			Name: "Error - ingredient already exist",
			IngDTO: dto.IngredientDTO{
				Name:        "Soul",
				Description: "",
				Quantity:    10,
			},
			CheckResult: true,
			ExpectedErr: errorList.ErrIngredientAlreadyExist,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			mockRepo := repoMock.NewAlchemyRepoInterface(t)

			switch tt.Name {
			case "Success":
				mockRepo.On("CheckIngredientExistsByName", context.Background(), tt.IngDTO.Name).Return(tt.CheckResult, tt.CheckErr)
				mockRepo.On("PostIngredients", context.Background(), tt.IngModel).Return(&tt.ExpectedResult, tt.ExpectedErr)
			case "Error - ingredient already exist":
				mockRepo.On("CheckIngredientExistsByName", context.Background(), tt.IngDTO.Name).Return(tt.CheckResult, tt.CheckErr)

			}

			serv := service.NewAlchemyService(mockRepo, nil)
			_, err := serv.PostIngredients(context.Background(), tt.IngDTO)

			if tt.ExpectedErr != nil {
				assert.ErrorIs(t, err, tt.ExpectedErr)
				return
			}
			assert.NoError(t, err)

		})

	}
}

func TestPostRecipe(t *testing.T) {
	tests := []struct {
		Name                       string
		CheckResult                bool
		CheckErr                   error
		RecipeDTO                  dto.RecipeDTO
		CreateRecipeErr            error
		ValueString                []string
		ValueArgs                  []any
		CreateRecipeIngredientsErr error
		ExpectedErr                error
	}{
		{
			Name:        "Success",
			CheckResult: false,
			CheckErr:    nil,
			RecipeDTO: dto.RecipeDTO{
				Name:               "Love elexir",
				BrewingTimeSeconds: 10,
				Ingredients: []dto.RecipeIngredientsDTO{
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
			CreateRecipeErr:            nil,
			ValueString:                []string{"($1, $2, $3)", "($4, $5, $6)"},
			ValueArgs:                  []any{0, 1, 10, 0, 2, 4},
			CreateRecipeIngredientsErr: nil,
			ExpectedErr:                nil,
		},
		{
			Name:        "Error - recipe already exist",
			CheckResult: true,
			CheckErr:    nil,
			RecipeDTO: dto.RecipeDTO{
				Name:               "Love elexir",
				BrewingTimeSeconds: 10,
			},
			ExpectedErr: errorList.ErrRecipeAlreadyExist,
		},
		{
			Name:        "Error - CheckExistError",
			CheckResult: false,
			CheckErr:    DB_ERROR,
			RecipeDTO: dto.RecipeDTO{
				Name:               "Love elexir",
				BrewingTimeSeconds: 10,
				Ingredients: []dto.RecipeIngredientsDTO{
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
			CreateRecipeErr:            nil,
			ValueString:                []string{"($1, $2, $3)", "($4, $5, $6)"},
			ValueArgs:                  []any{0, 1, 10, 0, 2, 4},
			CreateRecipeIngredientsErr: nil,
			ExpectedErr:                DB_ERROR,
		},
		{
			Name:        "Error - create recipe ingredients",
			CheckResult: false,
			CheckErr:    nil,
			RecipeDTO: dto.RecipeDTO{
				Name:               "Love elexir",
				BrewingTimeSeconds: 10,
				Ingredients: []dto.RecipeIngredientsDTO{
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
			CreateRecipeErr:            nil,
			ValueString:                []string{"($1, $2, $3)", "($4, $5, $6)"},
			ValueArgs:                  []any{0, 1, 10, 0, 2, 4},
			CreateRecipeIngredientsErr: DB_ERROR,
			ExpectedErr:                DB_ERROR,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			mockRepo := repoMock.NewAlchemyRepoInterface(t)
			mockTrans := transMock.NewTransactorInterface(t)
			switch tt.Name {
			case "Success":
				mockRepo.On("CheckExistRecipeByName", context.Background(), tt.RecipeDTO.Name).Return(tt.CheckResult, tt.CheckErr)
				mockTrans.On("WithinTransaction", context.Background(), mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(context.Background())
					}).
					Return(nil)
				mockRepo.On("CreateRecipe", context.Background(), &models.Recipe{Name: "Love elexir", BrewingTimeSeconds: 10}).Return(tt.CreateRecipeErr)
				mockRepo.On("CreateRecipeIngredients", context.Background(), tt.ValueString, tt.ValueArgs).Return(tt.CreateRecipeIngredientsErr)

			case "Error - recipe already exist":
				mockRepo.On("CheckExistRecipeByName", context.Background(), tt.RecipeDTO.Name).Return(tt.CheckResult, tt.CheckErr)

			case "Error - CheckExistError":
				mockRepo.On("CheckExistRecipeByName", context.Background(), tt.RecipeDTO.Name).Return(tt.CheckResult, tt.CheckErr)

			case "Error - create recipe ingredients":
				mockRepo.On("CheckExistRecipeByName", context.Background(), tt.RecipeDTO.Name).Return(tt.CheckResult, tt.CheckErr)
				mockTrans.On("WithinTransaction", context.Background(), mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(context.Background())
					}).
					Return(tt.CreateRecipeIngredientsErr)
				mockRepo.On("CreateRecipe", context.Background(), &models.Recipe{Name: "Love elexir", BrewingTimeSeconds: 10}).Return(tt.CreateRecipeErr)
				mockRepo.On("CreateRecipeIngredients", context.Background(), tt.ValueString, tt.ValueArgs).Return(tt.CreateRecipeIngredientsErr)
			}

			serv := service.NewAlchemyService(mockRepo, mockTrans)
			res, err := serv.PostRecipe(context.Background(), tt.RecipeDTO)

			if tt.ExpectedErr != nil {
				assert.ErrorIs(t, err, tt.ExpectedErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.RecipeDTO.Name, res.Name)
			assert.Equal(t, tt.RecipeDTO.BrewingTimeSeconds, res.BrewingTimeSeconds)
			assert.Equal(t, tt.RecipeDTO.Description, res.Description)
			for i, value := range res.Ingredients {
				assert.Equal(t, tt.RecipeDTO.Ingredients[i].IngredientID, value.IngredientID)
				assert.Equal(t, tt.RecipeDTO.Ingredients[i].QuantityNeeded, value.QuantityNeeded)
			}

		})

	}
}
