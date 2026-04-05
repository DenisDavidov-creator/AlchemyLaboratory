package main

import (
	"alla/worker-service/internal/boiler-worker/repository"
	"alla/worker-service/internal/boiler-worker/service"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	dto "alla/shared/DTO"
	"alla/shared/errorList"
	pb "alla/shared/pb"

	"github.com/subosito/gotenv"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger := newLogger()

	err := gotenv.Load()
	if err != nil {
		logger.Error("Failed get ENV",
			slog.String("err", err.Error()),
		)
		os.Exit(1)
	}

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
	defer kafkaClient.Close()

	DB_SERVICE_GRPC := os.Getenv("DB_SERVICE_GRPC")
	conn, err := grpc.NewClient(DB_SERVICE_GRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to connect to service by gRPC",
			slog.String("err", err.Error()),
		)
		os.Exit(1)
	}
	defer conn.Close()

	brewingClient := pb.NewJobServiceClient(conn)

	repoBrewing := repository.NewRepoBrewing(brewingClient)

	serviceBrewing := service.NewBoilerWorker(repoBrewing)

	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-quit
		logger.Info("Shouting down")
		cancel()
		kafkaClient.Close()
		conn.Close()
	}()

	logger.Info("Starting Kafka consumer")
	for {
		fetches := kafkaClient.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return
		}
		fetches.EachRecord(func(r *kgo.Record) {
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
		})
	}

}

func newLogger() *slog.Logger {
	var handler slog.Handler
	if os.Getenv("ENV") == "production" {
		opts := &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}
