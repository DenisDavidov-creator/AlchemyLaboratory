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

func TestCreateJob(t *testing.T) {
	tests := []struct {
		Name          string
		DTO           dto.JobDTO
		ModelRes      models.BrewingJobs
		WantError     bool
		ExpectedError error
		UUID          string
	}{
		{
			Name: "Success",
			DTO: dto.JobDTO{
				RecipeID: 1,
				Details:  "Something details",
			},
			WantError: false,
			UUID:      "UUID",
			ModelRes: models.BrewingJobs{
				RecipeID: 1,
				Status:   "queued",
				Details:  "Something details",
			},
		},
		{
			Name: "Failed",
			DTO: dto.JobDTO{
				RecipeID: 1,
				Details:  "Something details",
			},
			WantError: true,
			UUID:      "UUID",
			ModelRes: models.BrewingJobs{
				RecipeID: 1,
				Status:   "queued",
				Details:  "Something details",
			},
			ExpectedError: ERROR_DB,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			brewingMockRepo := mocks.NewBrewingRepoInterface(t)
			switch tt.Name {
			case "Success":
				brewingMockRepo.On("CreateJob", context.Background(), &tt.ModelRes).
					Run(func(args mock.Arguments) {
						out := args.Get(1).(*models.BrewingJobs)
						out.PublicID = "UUID"
					}).
					Return(tt.ExpectedError)
			case "Failed":
				brewingMockRepo.On("CreateJob", context.Background(), &tt.ModelRes).
					Run(func(args mock.Arguments) {
						out := args.Get(1).(*models.BrewingJobs)
						out.PublicID = "UUID"
					}).
					Return(tt.ExpectedError)
			}

			serv := service.NewBrewingService(brewingMockRepo, nil, nil)

			uuid, err := serv.CreateJob(context.Background(), tt.DTO)
			if tt.WantError {
				assert.ErrorIs(t, err, tt.ExpectedError)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.UUID, uuid.JobUUID)
		})
	}
}

func TestGetBrewStatus(t *testing.T) {
	tests := []struct {
		Name          string
		ExpectedError error
		UUID          dto.JobUUIDDTO
		Status        string
		res           dto.JobStatusresponseDTO
	}{
		{
			Name:          "Success",
			ExpectedError: nil,
			res: dto.JobStatusresponseDTO{
				Status: "queued",
			},
			UUID:   dto.JobUUIDDTO{JobUUID: "UUID"},
			Status: "queued",
		},
		{
			Name:          "Failed",
			ExpectedError: ERROR_DB,
			UUID:          dto.JobUUIDDTO{JobUUID: "UUID"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			brewingMockRepo := mocks.NewBrewingRepoInterface(t)
			switch tt.Name {
			case "Success":
				brewingMockRepo.On("GetBrewStatus", context.Background(), tt.UUID.JobUUID).Return(tt.Status, tt.ExpectedError)

			case "Failed":
				brewingMockRepo.On("GetBrewStatus", context.Background(), tt.UUID.JobUUID).Return("", tt.ExpectedError)
			}

			srv := service.NewBrewingService(brewingMockRepo, nil, nil)
			dto, err := srv.GetBrewStatus(context.Background(), tt.UUID)

			if tt.ExpectedError != nil {
				assert.ErrorIs(t, err, tt.ExpectedError)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.Status, dto.Status)
		})

	}
}

func TestSetStatus(t *testing.T) {
	tests := []struct {
		Name          string
		Req           dto.JobStatusDTO
		expectedError error
	}{
		{
			Name:          "Success",
			expectedError: nil,
			Req:           dto.JobStatusDTO{UUID: "UUID", Status: "queued"},
		},
		{
			Name:          "Failed",
			expectedError: ERROR_DB,
			Req:           dto.JobStatusDTO{UUID: "UUID", Status: "queued"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			brewingMockRepo := mocks.NewBrewingRepoInterface(t)
			switch tt.Name {
			case "Success":
				brewingMockRepo.On("SetStatus", context.Background(), tt.Req.UUID, tt.Req.Status).Return(tt.expectedError)
			case "Failed":
				brewingMockRepo.On("SetStatus", context.Background(), tt.Req.UUID, tt.Req.Status).Return(tt.expectedError)

			}

			srv := service.NewBrewingService(brewingMockRepo, nil, nil)
			err := srv.SetStatus(context.Background(), tt.Req)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				return
			}

			assert.NoError(t, err)

		})
	}
}
