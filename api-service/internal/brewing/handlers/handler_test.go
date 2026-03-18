package handlers_test

import (
	"alla/api-service/internal/brewing/handlers"
	"alla/api-service/internal/brewing/service/mocks"
	dto "alla/shared/DTO"
	"strings"

	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStatusBrew(t *testing.T) {

	tests := []struct {
		name           string
		uuidParam      string
		mockReturn     *dto.JobStatusresponseDTO
		mockErr        error
		expectedStatus int
	}{
		{name: "Success - Returns status",
			uuidParam: "fksjl-fsjdklfj-fsdjkldsjf-",
			mockReturn: &dto.JobStatusresponseDTO{
				Status: "queude",
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{name: "Fail - Database fail",
			uuidParam:      "fksjl-fsjdklfj-fsdjkldsjf-",
			mockReturn:     nil,
			mockErr:        errors.New("DB fail"),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewBrewingServiceInterface(t)
			handler := handlers.NewBrewingHandler(mockService)

			mockService.On("GetBrewStatus", mock.Anything, mock.Anything).Return(tt.mockReturn, tt.mockErr)

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
		mockRepoReturn *dto.JobUUIDDTO
		mockRepoErr    error
		expectedStatus int
	}{
		{name: "Success - Created",
			inputJSON: `{"recipe_id":10}`,
			mockRepoReturn: &dto.JobUUIDDTO{
				JobUUID: "I kissed a girl",
			},
			mockRepoErr:    nil,
			expectedStatus: http.StatusOK,
		}, {
			name:           "Bad JSON",
			inputJSON:      `{"recipe_id": "not-a-number"}`,
			mockRepoReturn: nil,
			mockRepoErr:    nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Validation Error - Zero ID",
			inputJSON:      `{"recipe_id": -1}`,
			mockRepoReturn: nil,
			mockRepoErr:    nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error - DB Fail on PostJob",
			inputJSON:      `{"recipe_id": 1}`,
			mockRepoReturn: nil,
			mockRepoErr:    errors.New("Db Fall"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewBrewingServiceInterface(t)
			handler := handlers.NewBrewingHandler(mockService)

			if tt.expectedStatus != http.StatusBadRequest {
				mockService.On("PostJob", mock.Anything, mock.AnythingOfType("dto.JobDTO")).Return(tt.mockRepoReturn, tt.mockRepoErr).Once()
			}

			req := httptest.NewRequest(http.MethodPost, "/brew", strings.NewReader(tt.inputJSON))
			w := httptest.NewRecorder()

			handler.Brew(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
