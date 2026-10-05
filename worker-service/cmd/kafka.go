package main

import (
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	"alla/worker-service/internal/boiler-worker/service"
	"alla/worker-service/metrics"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

var retryDelays = []time.Duration{5 * time.Second, 30 * time.Second, 2 * time.Minute}

func NewKafkaClient(logger *slog.Logger) (*kgo.Client, error) {
	kafkaClient, err := kgo.NewClient(
		kgo.SeedBrokers(os.Getenv("KAFKA_ADDR")),
		kgo.ConsumerGroup("worker-group"),
		kgo.ConsumeTopics("brew-jobs", "brew-jobs.retry"),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		logger.Error("Failed connect to Kafka",
			slog.String("err", err.Error()),
			slog.String("addr", os.Getenv("KAFKA_ADDR")),
		)
		return nil, fmt.Errorf("failed connect to kafka %w", err)
	}
	return kafkaClient, nil
}

func ProcessRecord(ctx context.Context, r *kgo.Record, kafkaClient *kgo.Client, serviceBrewing service.ServiceInterface, logger *slog.Logger) error {
	metrics.JobsInFlight.Inc()
	defer metrics.JobsInFlight.Dec()

	var msg struct {
		JobUUID string `json:"job_uuid"`
	}
	if err := json.Unmarshal(r.Value, &msg); err != nil {
		logger.Error("Unmarshal error",
			slog.String("err", err.Error()),
		)
		return nil
	}
	jobLogger := logger.With(
		slog.String("job_uuid", msg.JobUUID),
	)

	var retryCount int
	var err error
	for _, h := range r.Headers {
		if h.Key == "retry-count" {
			retryCount, err = strconv.Atoi(string(h.Value))
			if err != nil {
				jobLogger.Error("invalid retry-count header",
					slog.String("err", err.Error()),
				)
			}
		}
	}

	err = waitForRetry(ctx, r, retryCount)
	if err != nil {
		return err
	}
	jobLogger.Info("Started boiled")
	if err := serviceBrewing.Boiled(ctx, dto.JobUUIDDTO{JobUUID: msg.JobUUID}); err != nil {
		if ctx.Err() != nil {
			jobLogger.Info("Brewing interrupted by shutdown")
			return fmt.Errorf("interrupted by shutdown:%w", ctx.Err())
		}
		if errors.Is(err, errorList.ErrIngredientNotEnough) {
			jobLogger.Warn("Not enough ingredients:",
				slog.String("err", err.Error()),
			)
			metrics.MessageProcessedTotal.WithLabelValues("error").Inc()
			return nil
		}
		jobLogger.Warn("Boiled, unexpected error", slog.String("err", err.Error()))

		if retryCount >= len(retryDelays) {
			jobLogger.Error("max retries reached",
				slog.Int("retry_count", retryCount),
			)
			err := kafkaClient.ProduceSync(ctx, &kgo.Record{
				Topic: "brew-jobs.dlq",
				Value: r.Value,
			}).FirstErr()
			if err != nil {
				return fmt.Errorf("add to dlq: %w", err)
			}

			err = serviceBrewing.MarkFailed(ctx, dto.JobUUIDDTO{JobUUID: msg.JobUUID})
			if err != nil {
				jobLogger.Error("technical error, mark failed", slog.String("err", err.Error()))
			}
			metrics.DlqTotal.WithLabelValues().Inc()
			return nil
		}
		jobLogger.Warn("technical error, sending to retry",
			slog.Int("count-retry", retryCount+1),
		)
		err := kafkaClient.ProduceSync(ctx, &kgo.Record{
			Topic: "brew-jobs.retry",
			Value: r.Value,
			Headers: []kgo.RecordHeader{
				{Key: "retry-count", Value: []byte(strconv.Itoa(retryCount + 1))},
			},
		}).FirstErr()
		if err != nil {
			return fmt.Errorf("add count to retry topic: %w", err)
		}
		metrics.RetriesTotal.WithLabelValues().Inc()
		return nil
	}
	metrics.MessageProcessedTotal.WithLabelValues("success").Inc()
	return nil
}

func waitForRetry(ctx context.Context, r *kgo.Record, retryCount int) error {
	var wait time.Duration

	if retryCount > 0 && retryCount <= len(retryDelays) {
		wait = retryDelays[retryCount-1]
	}
	readyAt := r.Timestamp.Add(wait)
	delay := time.Until(readyAt)
	if delay <= 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
	}
	return nil
}
