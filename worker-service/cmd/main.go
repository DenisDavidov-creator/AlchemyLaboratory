package main

import (
	"alla/worker-service/internal/boiler-worker/repository"
	"alla/worker-service/internal/boiler-worker/service"
	"alla/worker-service/internal/pool"
	"alla/worker-service/metrics"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	pb "alla/shared/pb"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/subosito/gotenv"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger := NewLogger()

	err := gotenv.Load()
	if err != nil {
		logger.Error("Failed get ENV",
			slog.String("err", err.Error()),
		)
		os.Exit(1)
	}

	kafkaClient := NewKafkaClient(logger)
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":9092", nil)
	}()

	size := poolSize(logger)
	workerPool := pool.New(size)
	metrics.PoolSize.Set(float64(size))

	logger.Info("Starting Kafka consumer")
	for {
		fetches := kafkaClient.PollFetches(ctx)
		if ctx.Err() != nil || fetches.IsClientClosed() {
			break
		}

		start := time.Now()

		var batch sync.WaitGroup
		fetches.EachRecord(func(r *kgo.Record) {
			batch.Add(1)
			workerPool.Submit(func() {
				defer batch.Done()
				ProcessRecord(ctx, r, kafkaClient, serviceBrewing, logger)
			})
		})
		batch.Wait()

		metrics.BatchSize.Observe(float64(fetches.NumRecords()))
		metrics.BatchDuration.Observe(time.Since(start).Seconds())

		if err := kafkaClient.CommitUncommittedOffsets(ctx); err != nil {
			logger.Warn("Kafka commit",
				slog.String("err", err.Error()),
			)
		}
	}

	workerPool.Stop()
	logger.Info("Shutting down")
}

func poolSize(logger *slog.Logger) int {
	const defaultSize = 5

	raw := os.Getenv("WORKER_POOL_SIZE")
	if raw == "" {
		return defaultSize
	}

	size, err := strconv.Atoi(raw)
	if err != nil || size <= 0 {
		logger.Error("invalid WORKER_POOL_SIZE",
			slog.String("value", raw),
		)
		os.Exit(1)
	}
	return size
}
