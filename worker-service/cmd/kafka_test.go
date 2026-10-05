package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestWaitForRetry(t *testing.T) {
	const remaining = 50 * time.Millisecond

	tests := []struct {
		name          string
		retryCount    int
		age           time.Duration
		ctx           func() context.Context
		expectedError error
		minWait       time.Duration
	}{
		{
			name:       "No retry",
			retryCount: 0,
		},
		{
			name:       "Delay already passed",
			retryCount: 1,
			age:        time.Hour,
		},
		{
			name:       "Waits from record timestamp",
			retryCount: 1,
			age:        retryDelays[0] - remaining,
			minWait:    remaining,
		},
		{
			name:       "Last retry uses last delay",
			retryCount: len(retryDelays),
			age:        retryDelays[len(retryDelays)-1] - remaining,
			minWait:    remaining,
		},
		{
			name:       "Canceled while waiting",
			retryCount: 1,
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expectedError: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.ctx != nil {
				ctx = tt.ctx()
			}
			r := &kgo.Record{Timestamp: time.Now().Add(-tt.age)}

			start := time.Now()
			err := waitForRetry(ctx, r, tt.retryCount)
			elapsed := time.Since(start)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.GreaterOrEqual(t, elapsed, tt.minWait-5*time.Millisecond)
			assert.Less(t, elapsed, tt.minWait+time.Second)
		})
	}
}
