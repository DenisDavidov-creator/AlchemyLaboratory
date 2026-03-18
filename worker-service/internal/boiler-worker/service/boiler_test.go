package service_test

import (
	dto "alla/shared/DTO"
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
	tests := []struct {
		name          string
		uuid          dto.JobUUIDDTO
		getJobResult  *dto.JobTimeDTO
		getJobError   error
		setStatusErr  error
		expectedError error
	}{
		{
			name: "Success",
			uuid: dto.JobUUIDDTO{
				JobUUID: "I've been staring at the edge of the water",
			},
			getJobResult: &dto.JobTimeDTO{
				BrweingTime: 0,
			},
			getJobError:   nil,
			setStatusErr:  nil,
			expectedError: nil,
		},
		{
			name: "Error - get time",
			uuid: dto.JobUUIDDTO{
				JobUUID: "I've been staring at the edge of the water",
			},
			getJobError:   DB_ERROR,
			setStatusErr:  nil,
			expectedError: DB_ERROR,
		},
		{
			name: "Error - set status",
			uuid: dto.JobUUIDDTO{
				JobUUID: "I've been staring at the edge of the water",
			},
			getJobError:   nil,
			setStatusErr:  DB_ERROR,
			expectedError: DB_ERROR,
		},
	}

	for _, tt := range tests {
		mockRepo := mocks.NewRepositoryBrewingInterface(t)
		workerServ := service.NewBoilerWorker(mockRepo)

		switch tt.name {
		case "Success":
			mockRepo.On("GetJobByUUID", mock.Anything, tt.uuid).Return(tt.getJobResult, tt.getJobError).Once()
			mockRepo.On("SetStatus", mock.Anything, dto.JobStatusDTO{UUID: tt.uuid.JobUUID, Status: "Completed"}).Return(tt.setStatusErr)
		case "Error - get time":
			mockRepo.On("GetJobByUUID", mock.Anything, tt.uuid).Return(tt.getJobResult, tt.getJobError).Once()
			mockRepo.On("SetStatus", mock.Anything, dto.JobStatusDTO{UUID: tt.uuid.JobUUID, Status: "failed"}).Return(tt.setStatusErr)
		case "Error - set status":
			mockRepo.On("GetJobByUUID", mock.Anything, tt.uuid).Return(tt.getJobResult, tt.getJobError).Once()
			mockRepo.On("SetStatus", mock.Anything, dto.JobStatusDTO{UUID: tt.uuid.JobUUID, Status: "Completed"}).Return(tt.setStatusErr)
		}

		err := workerServ.Boiled(context.Background(), tt.uuid)

		if tt.expectedError != nil {
			assert.ErrorIs(t, err, tt.expectedError)
			return
		}

		assert.NoError(t, err)

	}
}
