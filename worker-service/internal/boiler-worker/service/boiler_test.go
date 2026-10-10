package service_test

import (
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	"alla/shared/status"
	"alla/worker-service/internal/boiler-worker/repository/mocks"
	"alla/worker-service/internal/boiler-worker/service"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var DB_ERROR error = errors.New("DB_ERROR")

func TestBoiled(t *testing.T) {
	uuid := dto.JobUUIDDTO{
		JobUUID: "I've been staring at the edge of the water",
	}

	tests := []struct {
		name          string
		ctx           func() context.Context
		setupMock     func(m *mocks.RepositoryBrewingInterface)
		expectedError error
	}{
		{
			name: "Success",
			setupMock: func(m *mocks.RepositoryBrewingInterface) {
				m.On("StartBrewing", mock.Anything, uuid).
					Return(&dto.JobTimeDTO{BrweingTime: 0}, nil).Once()
				m.On("SetStatus", mock.Anything, dto.JobStatusDTO{UUID: uuid.JobUUID, Status: status.StatusCompleted}).
					Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "Error - get time",
			setupMock: func(m *mocks.RepositoryBrewingInterface) {
				m.On("StartBrewing", mock.Anything, uuid).
					Return(nil, DB_ERROR).Once()
			},
			expectedError: DB_ERROR,
		},
		{
			name: "Error - ingredient not enough",
			setupMock: func(m *mocks.RepositoryBrewingInterface) {
				m.On("StartBrewing", mock.Anything, uuid).
					Return(&dto.JobTimeDTO{BrweingTime: 0}, errorList.ErrIngredientNotEnough).Once()
				m.On("SetStatus", mock.Anything, dto.JobStatusDTO{UUID: uuid.JobUUID, Status: status.StatusFailed}).
					Return(nil).Once()
			},
			expectedError: errorList.ErrIngredientNotEnough,
		},
		{
			name: "Error - set status",
			setupMock: func(m *mocks.RepositoryBrewingInterface) {
				m.On("StartBrewing", mock.Anything, uuid).
					Return(&dto.JobTimeDTO{BrweingTime: 0}, nil).Once()
				m.On("SetStatus", mock.Anything, dto.JobStatusDTO{UUID: uuid.JobUUID, Status: status.StatusCompleted}).
					Return(DB_ERROR).Once()
			},
			expectedError: DB_ERROR,
		},
		{
			name: "Canceled while brewing",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			setupMock: func(m *mocks.RepositoryBrewingInterface) {
				m.On("StartBrewing", mock.Anything, uuid).
					Return(&dto.JobTimeDTO{BrweingTime: 3600}, nil).Once()
			},
			expectedError: context.Canceled,
		},
		{
			name: "brewer dead",
			setupMock: func(m *mocks.RepositoryBrewingInterface) {
				m.On("StartBrewing", mock.Anything, uuid).
					Return(&dto.JobTimeDTO{BrweingTime: 0, Resumed: true}, nil).Once()
				m.On("SetStatus", mock.Anything, dto.JobStatusDTO{UUID: uuid.JobUUID, Status: status.StatusFailed}).
					Return(nil).Once()
			},
			expectedError: errorList.ErrInvalidStatusTransition,
		},
		{
			name: "brewer completed",
			setupMock: func(m *mocks.RepositoryBrewingInterface) {
				m.On("StartBrewing", mock.Anything, uuid).
					Return(&dto.JobTimeDTO{BrweingTime: 0, Resumed: true}, nil).Once()
				m.On("SetStatus", mock.Anything, dto.JobStatusDTO{UUID: uuid.JobUUID, Status: status.StatusFailed}).
					Return(errorList.ErrInvalidStatusTransition).Once()
			},
			expectedError: errorList.ErrInvalidStatusTransition,
		},
		{
			name: "brewer db-error",
			setupMock: func(m *mocks.RepositoryBrewingInterface) {
				m.On("StartBrewing", mock.Anything, uuid).
					Return(&dto.JobTimeDTO{BrweingTime: 0, Resumed: true}, nil).Once()
				m.On("SetStatus", mock.Anything, dto.JobStatusDTO{UUID: uuid.JobUUID, Status: status.StatusFailed}).
					Return(DB_ERROR).Once()
			},
			expectedError: DB_ERROR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewRepositoryBrewingInterface(t)
			workerServ := service.NewBoilerWorker(mockRepo, 0)

			tt.setupMock(mockRepo)

			ctx := context.Background()
			if tt.ctx != nil {
				ctx = tt.ctx()
			}

			err := workerServ.Boiled(ctx, uuid)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				return
			}

			assert.NoError(t, err)
		})
	}
}
