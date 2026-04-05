package main

import (
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	"alla/worker-service/internal/boiler-worker/service"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strconv"

	"github.com/twmb/franz-go/pkg/kgo"
)

func NewKafkaClient(logger *slog.Logger) *kgo.Client {
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
		os.Exit(1)
	}
	return kafkaClient
}

func ProcessRecord(ctx context.Context, r *kgo.Record, kafkaClient *kgo.Client, serviceBrewing service.ServiceInterface, logger *slog.Logger) {
	var msg struct {
		JobUUID string `json:"job_uuid"`
	}
	if err := json.Unmarshal(r.Value, &msg); err != nil {
		logger.Error("Unmarshal error",
			slog.String("err", err.Error()),
		)
		return
	}
	jobLogger := logger.With(
		slog.String("job_uuid", msg.JobUUID),
	)

	jobLogger.Info("Started boiled")
	if err := serviceBrewing.Boiled(context.Background(), dto.JobUUIDDTO{JobUUID: msg.JobUUID}); err != nil {
		if errors.Is(err, errorList.ErrIngredientNotEnough) {
			jobLogger.Warn("Not enough ingredients:",
				slog.String("err", err.Error()),
			)
			if err := kafkaClient.CommitRecords(ctx, r); err != nil {
				logger.Warn("Kafka commit",
					"err", err.Error(),
				)
			}
			return
		}
		jobLogger.Warn("Boiled, unexpected error", slog.String("err", err.Error()))

		var retryCount int
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
		if retryCount >= 3 {
			jobLogger.Error("max retries reached",
				slog.Int("retry_count", retryCount),
			)
			kafkaClient.ProduceSync(ctx, &kgo.Record{
				Topic: "brew-jobs.dlq",
				Value: r.Value,
			})
			return
		}
		jobLogger.Warn("Technical error, sending to retry",
			slog.Int("count-retry", retryCount+1),
		)
		kafkaClient.ProduceSync(ctx, &kgo.Record{
			Topic: "brew-jobs.retry",
			Value: r.Value,
			Headers: []kgo.RecordHeader{
				{Key: "retry-count", Value: []byte(strconv.Itoa(retryCount + 1))},
			},
		})
		return
	}
	if err := kafkaClient.CommitRecords(ctx, r); err != nil {
		jobLogger.Warn("Kafka commit",
			slog.String("err", err.Error()),
		)

	}
}
