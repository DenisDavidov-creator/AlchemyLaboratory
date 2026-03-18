package handlers_test

import (
	"alla/api-service/internal/alchemy/handlers"
	"alla/api-service/internal/alchemy/service/mocks"
	dto "alla/shared/DTO"
	errorList "alla/shared/errorList"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// type MockRepo struct {
// 	mock.Mock
// }
// func (mr *MockRepo) PostIngredients(ctx context.Context, m models.Ingredient) (*models.Ingredient, error) {
// 	args := mr.Called(ctx, m)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(*models.Ingredient), args.Error(1)
// }

func TestShowRecipe(t *testing.T) {
	tests := []struct {
		name           string
		mockReturn     []dto.RecipeResponseDTO
		mockErr        error
		expectedStatus int
	}{
		{name: "Success - Returns Recipes",
			mockReturn: []dto.RecipeResponseDTO{
				{
					ID:                 1,
					Name:               "Health Potion",
					BrewingTimeSeconds: 10,
					Ingredients: []dto.RecipeIngredientsDTO{
						{
							IngredientID:   1,
							QuantityNeeded: 10,
						},
						{
							IngredientID:   1,
							QuantityNeeded: 10,
						},
					},
				},
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{name: "Database Failure",
			mockReturn:     []dto.RecipeResponseDTO{},
			mockErr:        errors.New("db down"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewServiceInterface(t)
			handler := handlers.NewGuildHandler(mockService)

			mockService.On("GetRecipes", mock.Anything).Return(tt.mockReturn, tt.mockErr)

			req := httptest.NewRequest("get", "/recipes", nil)
			w := httptest.NewRecorder()

			handler.ShowRecipes(w, req)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, http.StatusOK, w.Code)
				var response []dto.RecipeResponseDTO
				json.Unmarshal(w.Body.Bytes(), &response)

				assert.Equal(t, tt.mockReturn[0].BrewingTimeSeconds, response[0].BrewingTimeSeconds)
				assert.Equal(t, tt.mockReturn[0].ID, response[0].ID)
				assert.Equal(t, tt.mockReturn[0].Name, response[0].Name)
				for i := range tt.mockReturn[0].Ingredients {
					assert.Equal(t, tt.mockReturn[0].Ingredients[i].IngredientID, response[0].Ingredients[i].IngredientID)
					assert.Equal(t, tt.mockReturn[0].Ingredients[i].QuantityNeeded, response[0].Ingredients[i].QuantityNeeded)
				}
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestBuyNewIngredients(t *testing.T) {
	tests := []struct {
		name           string
		inputJSON      string
		mockReturn     *dto.IngredientResponseDTO
		mockErr        error
		expectedStatus int
	}{
		{name: "Success - Created",
			inputJSON: `{"name": "Dragon Scale", "description":"Shiny", "quantity":10}`,
			mockReturn: &dto.IngredientResponseDTO{
				ID: 1, Name: "Dragon Scale", Description: "Shiny", Quantity: 10,
			},
			mockErr:        nil,
			expectedStatus: http.StatusCreated,
		},
		{name: "Error - Invalid JSON Body",
			inputJSON:      `{"name" :"Bodfod"`,
			mockReturn:     nil,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{name: "Error - Not enough parameters",
			inputJSON:      `{"name" :"Bodfod", "description":"Shiny"}`,
			mockReturn:     nil,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{name: "Error - negative quantity",
			inputJSON:      `{"name": "Dragon Scale", "description":"Shiny", "quantity":-1}`,
			mockReturn:     nil,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{name: "Error - Db fail",
			inputJSON:      `{"name": "Dragon Scale", "description":"Shiny", "quantity":1}`,
			mockReturn:     nil,
			mockErr:        errors.New("Db fail"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewServiceInterface(t)
			handler := handlers.NewGuildHandler(mockService)

			if tt.expectedStatus == http.StatusCreated || tt.name == "Error - Db fail" {
				mockService.On("PostIngredients", mock.Anything, mock.AnythingOfType("dto.IngredientDTO")).
					Return(tt.mockReturn, tt.mockErr)
			}

			req := httptest.NewRequest(http.MethodPost, "/ingredients", strings.NewReader(tt.inputJSON))
			w := httptest.NewRecorder()

			handler.BuyNewIngredients(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				mockService.AssertExpectations(t)
				var response dto.IngredientResponseDTO
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, response.ID, tt.mockReturn.ID)
				assert.Equal(t, response.Name, tt.mockReturn.Name)
			}
		})
	}
}

func TestShowIngredients(t *testing.T) {

	tests := []struct {
		name           string
		expectedResult []dto.IngredientResponseDTO
		expectedErr    error
		expectedStatus int
	}{
		{
			name: "Success - Shows",
			expectedResult: []dto.IngredientResponseDTO{
				{ID: 1, Name: "Dragonfly", Quantity: 10},
			},
			expectedErr:    nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Error - DB fail",
			expectedResult: nil,
			expectedErr:    errors.New("DB fail"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		mockService := mocks.NewServiceInterface(t)
		handler := handlers.NewGuildHandler(mockService)

		mockService.On("GetIngredients", mock.Anything).Return(tt.expectedResult, tt.expectedErr)

		req := httptest.NewRequest("GET", "/ingredients", nil)
		w := httptest.NewRecorder()

		handler.ShowIngredients(w, req)

		if tt.expectedStatus == http.StatusOK {
			assert.Equal(t, http.StatusOK, w.Code)
			var response []dto.IngredientResponseDTO
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(t, tt.expectedResult, response)
		}
		mockService.AssertExpectations(t)
	}

}

func TestBuyExistingIngredients(t *testing.T) {
	tests := []struct {
		name           string
		idParam        string
		inputJSON      string
		mockErr        error
		expectedStatus int
	}{
		{name: "Success - Created",
			idParam:        `10`,
			inputJSON:      `{"quantity": 5, "name":"herbs"}`,
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{name: "Negative quantity",
			idParam:        `10`,
			inputJSON:      `{"quantity": -5}`,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{name: "Wrong idParam",
			idParam:        `asd`,
			inputJSON:      `{"quantity": 5}`,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{name: "Wrong quantity",
			idParam:        `asd`,
			inputJSON:      `{"quantity": asdsad}`,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{name: "Error - Wrong json format",
			idParam:        `10`,
			inputJSON:      `{"quantity": "10"}`,
			mockErr:        errorList.ErrWrongJsonFormat,
			expectedStatus: http.StatusBadRequest,
		},
		{name: "DB fail",
			idParam:        `10`,
			inputJSON:      `{"quantity": 10}`,
			mockErr:        errors.New("Db fail"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewServiceInterface(t)
			handler := handlers.NewGuildHandler(mockService)

			if tt.expectedStatus != http.StatusBadRequest || tt.name == "DB fail" {
				mockService.On("AddIngredients", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockErr)
			}

			req := httptest.NewRequest(http.MethodPost, "/ingredients/"+tt.idParam, strings.NewReader(tt.inputJSON))
			req = mux.SetURLVars(req, map[string]string{"id": tt.idParam})

			w := httptest.NewRecorder()

			handler.BuyExistIngredient(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			mockService.AssertExpectations(t)
		})
	}
}

func TestCreateRecipe(t *testing.T) {
	tests := []struct {
		name           string
		inputJSON      string
		mockRepoReturn *dto.RecipeResponseDTO
		mockRepoErr    error
		expectedStatus int
	}{
		{name: "Success - Created",
			inputJSON: `{"name":"grass", "brewing_time_seconds": 50, "ingredients":[
				{
					"ingredient_id":1,
					"quantity_needed":3
				}
			] }`,
			mockRepoReturn: &dto.RecipeResponseDTO{
				ID:                 1,
				Name:               "grass",
				BrewingTimeSeconds: 50,
				Ingredients: []dto.RecipeIngredientsDTO{
					{
						IngredientID:   1,
						QuantityNeeded: 3,
					},
				},
			},
			mockRepoErr:    nil,
			expectedStatus: http.StatusCreated,
		},
		{name: "Error - don't have ingredients",
			inputJSON:      `{"name":"grass", "brewing_time_seconds": 50, "ingredients":[] }`,
			mockRepoReturn: nil,
			mockRepoErr:    nil,
			expectedStatus: http.StatusBadRequest,
		},
		{name: "Error - Wrong json format",
			inputJSON: `{"name":"grass", "brewing_time_seconds": "asd", "ingredients":[
			{
					"ingredient_id":1,
					"quantity_needed":3
				}
			] }`,
			mockRepoReturn: nil,
			mockRepoErr:    nil,
			expectedStatus: http.StatusBadRequest,
		},
		{name: "Error - DB fail",
			inputJSON: `{"name":"grass", "brewing_time_seconds": 50, "ingredients":[
				{
					"ingredient_id":1,
					"quantity_needed":3
				}
			] }`,
			mockRepoReturn: nil,
			mockRepoErr:    errors.New("Db fail"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewServiceInterface(t)
			handler := handlers.NewGuildHandler(mockService)

			if tt.expectedStatus != http.StatusBadRequest {
				mockService.On("PostRecipe", mock.Anything, mock.AnythingOfType("dto.RecipeDTO")).Once().
					Return(tt.mockRepoReturn, tt.mockRepoErr)
			}

			req := httptest.NewRequest(http.MethodPost, "/recipes", strings.NewReader(tt.inputJSON))
			w := httptest.NewRecorder()

			handler.CreateRecipe(w, req)

			if tt.expectedStatus == http.StatusCreated {
				var response dto.RecipeResponseDTO
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, *tt.mockRepoReturn, response)

			}
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
