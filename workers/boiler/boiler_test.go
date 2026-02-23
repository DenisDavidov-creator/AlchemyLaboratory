package boiler_test

import (
	"alchemicallabaratory/repository/mocks"
	"alchemicallabaratory/workers/boiler"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBoiled(t *testing.T) {
	tests := []struct {
		name               string
		uuid               string
		expectedResult     int
		expectedSecondsErr error
		expectedStatusErr  error
	}{
		{
			name:               "Correct - Ok",
			uuid:               "I-want-to-be-a-girl",
			expectedResult:     0,
			expectedSecondsErr: nil,
			expectedStatusErr:  nil,
		},
		{
			name:               "Error - get seconds",
			uuid:               "I-want-to-be-a-girl",
			expectedResult:     0,
			expectedSecondsErr: errors.New("db fail"),
			expectedStatusErr:  nil,
		},
		{
			name:               "Error - set status",
			uuid:               "I-want-to-be-a-girl",
			expectedResult:     0,
			expectedSecondsErr: nil,
			expectedStatusErr:  errors.New("db fail "),
		},
	}

	for _, tt := range tests {
		mockRepo := mocks.NewGrimoireRepoInterface(t)
		worker := boiler.NewBoilerWorker(mockRepo)

		mockRepo.On("GetJobByUUID", mock.Anything, mock.Anything).Return(tt.expectedResult, tt.expectedSecondsErr).Once()

		if tt.expectedSecondsErr == nil {
			mockRepo.On("SetStatus", mock.Anything, mock.Anything, mock.Anything).Return(tt.expectedStatusErr)
		}

		err := worker.Boiled(context.Background(), tt.uuid)
		if tt.expectedSecondsErr == nil && tt.expectedStatusErr == nil {
			assert.NoError(t, err)
		}
	}
}
