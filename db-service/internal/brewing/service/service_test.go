package service_test

import (
	alchemyRepoMock "alla/db-service/internal/alchemy/repository/mocks"
	"alla/db-service/internal/brewing/repository/mocks"
	"alla/db-service/internal/brewing/service"
	"alla/db-service/models"
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	transMock "alla/db-service/internal/transactor/mocks"
)

var ERROR_DB error = errors.New("ERROR_DB")

func TestStartBrewing(t *testing.T) {
	tests := []struct {
		Name          string
		expectedError error
		ings          []models.RecipeIngredients
		brewingTime   int
		UUID          dto.JobUUIDDTO
	}{
		{
			Name:          "Success",
			expectedError: nil,
			ings: []models.RecipeIngredients{
				{
					IngredientID:   1,
					QuantityNeeded: 2,
				},
				{
					IngredientID:   2,
					QuantityNeeded: 4,
				},
			},
			brewingTime: 10,
			UUID:        dto.JobUUIDDTO{JobUUID: "I want to be a cat"},
		},
		{
			Name:          "Error - recipe not found",
			expectedError: errorList.ErrRecipeNotFound,
			UUID:          dto.JobUUIDDTO{JobUUID: "I want to be a cat"},
		},
		{
			Name:          "Error - not enough ingredients",
			expectedError: errorList.ErrIngredientNotEnough,
			ings: []models.RecipeIngredients{
				{
					IngredientID:   1,
					QuantityNeeded: 2,
				},
				{
					IngredientID:   2,
					QuantityNeeded: 4,
				},
			},
			brewingTime: 10,
			UUID:        dto.JobUUIDDTO{JobUUID: "I want to be a cat"},
		},
		{
			Name:          "Error - wrong set status",
			expectedError: ERROR_DB,
			ings: []models.RecipeIngredients{
				{
					IngredientID:   1,
					QuantityNeeded: 2,
				},
				{
					IngredientID:   2,
					QuantityNeeded: 4,
				},
			},
			brewingTime: 10,
			UUID:        dto.JobUUIDDTO{JobUUID: "I want to be a cat"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			alchemyMockRepo := alchemyRepoMock.NewAlchemyRepoInterface(t)
			brewingMockRepo := mocks.NewBrewingRepoInterface(t)
			mockTrans := transMock.NewTransactorInterface(t)

			switch tt.Name {
			case "Success":
				alchemyMockRepo.On("GetRecipeID", context.Background(), tt.UUID.JobUUID).Return(1, nil)
				alchemyMockRepo.On("GetIngredientsByRecipe", context.Background(), 1).Return(tt.ings, nil)
				alchemyMockRepo.On("GetBrewingTime", context.Background(), 1).Return(tt.brewingTime, nil)
				mockTrans.On("WithinTransaction", context.Background(), mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(context.Background())
					}).
					Return(nil)
				alchemyMockRepo.On("CheckingIngridients", context.Background(), tt.ings).Return(nil)
				brewingMockRepo.On("SetStatus", context.Background(), tt.UUID.JobUUID, "queued").Return(nil)
			case "Error - recipe not found":
				alchemyMockRepo.On("GetRecipeID", context.Background(), tt.UUID.JobUUID).Return(1, errorList.ErrRecipeNotFound)

			case "Error - not enough ingredients":
				alchemyMockRepo.On("GetRecipeID", context.Background(), tt.UUID.JobUUID).Return(1, nil)
				alchemyMockRepo.On("GetIngredientsByRecipe", context.Background(), 1).Return(tt.ings, nil)
				alchemyMockRepo.On("GetBrewingTime", context.Background(), 1).Return(tt.brewingTime, nil)
				mockTrans.On("WithinTransaction", context.Background(), mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(context.Background())
					}).
					Return(errorList.ErrIngredientNotEnough)
				alchemyMockRepo.On("CheckingIngridients", context.Background(), tt.ings).Return(errorList.ErrIngredientNotEnough)

			case "Error - wrong set status":
				alchemyMockRepo.On("GetRecipeID", context.Background(), tt.UUID.JobUUID).Return(1, nil)
				alchemyMockRepo.On("GetIngredientsByRecipe", context.Background(), 1).Return(tt.ings, nil)
				alchemyMockRepo.On("GetBrewingTime", context.Background(), 1).Return(tt.brewingTime, nil)
				mockTrans.On("WithinTransaction", context.Background(), mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(context.Background())
					}).
					Return(ERROR_DB)
				alchemyMockRepo.On("CheckingIngridients", context.Background(), tt.ings).Return(nil)
				brewingMockRepo.On("SetStatus", context.Background(), tt.UUID.JobUUID, "queued").Return(ERROR_DB)
			}

			serv := service.NewBrewingService(brewingMockRepo, alchemyMockRepo, mockTrans)

			brewingTime, err := serv.StartBrewing(context.Background(), tt.UUID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.brewingTime, brewingTime.BrweingTime)
		})

	}
}
