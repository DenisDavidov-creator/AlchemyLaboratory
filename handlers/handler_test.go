package handlers_test

import (
	"alchemicallabaratory/errorList"
	"alchemicallabaratory/handlers"
	"alchemicallabaratory/models"
	mocksG "alchemicallabaratory/repository/mocks"
	"alchemicallabaratory/workers/boiler"
	mocksB "alchemicallabaratory/workers/boiler/mocks"
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

func TestShowIngredients(t *testing.T) {

	tests := []struct {
		name           string
		expectedResult []models.Ingredient
		expectedErr    error
		expectedStatus int
	}{
		{
			name: "Success - Shows",
			expectedResult: []models.Ingredient{
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
		mockRepo := mocksG.NewGrimoireRepoInterface(t)
		handler := handlers.NewGuildHandler(mockRepo, boiler.NewBoilerWorker(mockRepo))

		mockRepo.On("GetIngredients", mock.Anything).Return(tt.expectedResult, tt.expectedErr)

		req := httptest.NewRequest("GET", "/ingredients", nil)
		w := httptest.NewRecorder()

		handler.ShowIngredients(w, req)

		if tt.expectedStatus == http.StatusOK {
			assert.Equal(t, http.StatusOK, w.Code)
			var response []models.Ingredient
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(t, tt.expectedResult, response)
		}
		mockRepo.AssertExpectations(t)
	}

}

func TestShowRecipe(t *testing.T) {
	tests := []struct {
		name           string
		mockReturn     []models.Recipe
		mockErr        error
		expectedStatus int
	}{
		{name: "Success - Returns Recipes",
			mockReturn: []models.Recipe{
				{ID: 1, Name: "Health Potion"},
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{name: "Database Failure",
			mockReturn:     []models.Recipe{},
			mockErr:        errors.New("db down"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocksG.NewGrimoireRepoInterface(t)
			handler := handlers.NewGuildHandler(mockRepo, nil)

			mockRepo.On("GetRecipes", mock.Anything).Return(tt.mockReturn, tt.mockErr)

			req := httptest.NewRequest("get", "/recipes", nil)
			w := httptest.NewRecorder()

			handler.ShowRecipes(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			mockRepo.AssertExpectations(t)
		})
	}

}

func TestBuyNewIngredients(t *testing.T) {
	tests := []struct {
		name           string
		inputJSON      string
		mockReturn     *models.Ingredient
		mockErr        error
		expectedStatus int
	}{
		{name: "Success - Created",
			inputJSON: `{"name": "Dragon Scale", "description":"Shiny", "quantity":10}`,
			mockReturn: &models.Ingredient{
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
			mockRepo := mocksG.NewGrimoireRepoInterface(t)
			handler := handlers.NewGuildHandler(mockRepo, nil)

			if tt.expectedStatus == http.StatusCreated || tt.name == "Error - Db fail" {
				mockRepo.On("PostIngredients", mock.Anything, mock.AnythingOfType("models.Ingredient")).
					Return(tt.mockReturn, tt.mockErr)
			}

			req := httptest.NewRequest(http.MethodPost, "/ingredients", strings.NewReader(tt.inputJSON))
			w := httptest.NewRecorder()

			handler.BuyNewIngredients(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				mockRepo.AssertExpectations(t)
				var response models.Ingredient
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, response.ID, tt.mockReturn.ID)
				assert.Equal(t, response.Name, tt.mockReturn.Name)
			}
		})
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
			mockRepo := mocksG.NewGrimoireRepoInterface(t)
			handler := handlers.NewGuildHandler(mockRepo, nil)

			if tt.expectedStatus != http.StatusBadRequest || tt.name == "DB fail" {
				mockRepo.On("AddIngredients", mock.Anything, mock.Anything, mock.Anything).
					Return(tt.mockErr)
			}

			req := httptest.NewRequest(http.MethodPost, "/ingredients/"+tt.idParam, strings.NewReader(tt.inputJSON))
			req = mux.SetURLVars(req, map[string]string{"id": tt.idParam})

			w := httptest.NewRecorder()

			handler.BuyExistIngredient(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestStatusBrew(t *testing.T) {

	tests := []struct {
		name           string
		uuidParam      string
		mockReturn     string
		mockErr        error
		expectedStatus int
	}{
		{name: "Success - Returns status",
			uuidParam:      "fksjl-fsjdklfj-fsdjkldsjf-",
			mockReturn:     "queued",
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{name: "Fail - Database fail",
			uuidParam:      "fksjl-fsjdklfj-fsdjkldsjf-",
			mockReturn:     "",
			mockErr:        errors.New("DB fail"),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocksG.NewGrimoireRepoInterface(t)
			handler := handlers.NewGuildHandler(mockRepo, nil)

			mockRepo.On("GetBrewStatus", mock.Anything, mock.Anything).Return(tt.mockReturn, tt.mockErr)

			req := httptest.NewRequest("get", "/brew/status/"+tt.uuidParam, nil)
			req = mux.SetURLVars(req, map[string]string{"uuid": tt.uuidParam})
			w := httptest.NewRecorder()

			handler.StatusBrew(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

		})

	}
}

func TestBrew(t *testing.T) {
	tests := []struct {
		name           string
		inputJSON      string
		mockRepoReturn *models.BrewingJobs
		mockRepoErr    error
		mockBoilErr    error
		expectedStatus int
	}{
		{name: "Success - Created",
			inputJSON: `{"recipe_id":10}`,
			mockRepoReturn: &models.BrewingJobs{
				ID:       1,
				PublicID: "sdfds-sdfds-sdfsf-sdfsdf",
				Status:   "completed",
				RecipeID: 10,
			},
			mockRepoErr:    nil,
			mockBoilErr:    nil,
			expectedStatus: http.StatusOK,
		}, {
			name:           "Bad JSON",
			inputJSON:      `{"recipe_id": "not-a-number"}`,
			mockRepoReturn: nil,
			mockRepoErr:    nil,
			mockBoilErr:    nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Validation Error - Zero ID",
			inputJSON:      `{"recipe_id": -1}`,
			mockRepoReturn: nil,
			mockRepoErr:    nil,
			mockBoilErr:    nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error - DB Fail on PostJob",
			inputJSON:      `{"recipe_id": 1}`,
			mockRepoReturn: nil,
			mockRepoErr:    errors.New("Db Fall"),
			mockBoilErr:    nil,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:      "Error - Boil Fail",
			inputJSON: `{"recipe_id": 10}`,
			mockRepoReturn: &models.BrewingJobs{
				ID:       1,
				PublicID: "sdfds-sdfds-sdfsf-sdfsdf",
				Status:   "queued",
				RecipeID: 10,
			},
			mockRepoErr:    nil,
			mockBoilErr:    errors.New("Boil Fail"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocksG.NewGrimoireRepoInterface(t)
			mockBoil := mocksB.NewBoilerWorkerInterface(t)
			handler := handlers.NewGuildHandler(mockRepo, mockBoil)

			if tt.expectedStatus != http.StatusBadRequest {
				// We use MatchedBy to ignore the CreatedAt field during comparison
				mockRepo.On("PostJob", mock.Anything, mock.MatchedBy(func(job models.BrewingJobs) bool {
					return job.RecipeID != 0 && job.Status == "queued"
				})).Return(tt.mockRepoReturn, tt.mockRepoErr).Once()
			}

			if tt.mockRepoErr == nil && tt.mockRepoReturn != nil {
				mockBoil.On("Boiled", mock.Anything, tt.mockRepoReturn.PublicID).
					Return(tt.mockBoilErr).Once()
			}

			req := httptest.NewRequest(http.MethodPost, "/brew", strings.NewReader(tt.inputJSON))
			w := httptest.NewRecorder()

			handler.Brew(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCreateRecipe(t *testing.T) {
	tests := []struct {
		name           string
		inputJSON      string
		mockRepoReturn *models.Recipe
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
			mockRepoReturn: &models.Recipe{
				ID:                 1,
				Name:               "grass",
				BrewingTimeSeconds: 50,
				Ingredients: []models.RecipeIngredients{
					{
						RecipeID:       1,
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
			mockRepo := mocksG.NewGrimoireRepoInterface(t)
			handler := handlers.NewGuildHandler(mockRepo, nil)

			if tt.expectedStatus != http.StatusBadRequest {
				mockRepo.On("PostRecipe", mock.Anything, mock.AnythingOfType("models.Recipe")).Once().
					Return(tt.mockRepoReturn, tt.mockRepoErr)
			}

			req := httptest.NewRequest(http.MethodPost, "/recipes", strings.NewReader(tt.inputJSON))
			w := httptest.NewRecorder()

			handler.CreateRecipe(w, req)

			if tt.expectedStatus == http.StatusCreated {
				var response models.Recipe
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, *tt.mockRepoReturn, response)

			}
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
