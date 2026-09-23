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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStatusBrew(t *testing.T) {

	const validUUID = "6f1e2a3b-4c5d-4e6f-8a9b-0c1d2e3f4a5b"

	tests := []struct {
		name           string
		uuidParam      string
		mockReturn     *dto.JobStatusresponseDTO
		mockErr        error
		expectedStatus int
	}{
		{name: "Success - Returns status",
			uuidParam: validUUID,
			mockReturn: &dto.JobStatusresponseDTO{
				Status: "queued",
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{name: "Fail - Database fail",
			uuidParam:      validUUID,
			mockReturn:     nil,
			mockErr:        errors.New("DB fail"),
			expectedStatus: http.StatusInternalServerError,
		},
		{name: "Fail - Invalid uuid",
			uuidParam:      "not-a-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{name: "Fail - Missing uuid",
			uuidParam:      "",
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewBrewingServiceInterface(t)
			handler := handlers.NewBrewingHandler(mockService)

			if tt.expectedStatus != http.StatusBadRequest {
				mockService.On("GetBrewStatus", mock.Anything, dto.JobUUIDDTO{JobUUID: tt.uuidParam}).Return(tt.mockReturn, tt.mockErr).Once()
			}

			req := httptest.NewRequest(http.MethodGet, "/brew/status?uuid="+tt.uuidParam, nil)
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
