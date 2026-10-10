package main

import (
	"alla/worker-service/internal/boiler-worker/repository"
	"alla/worker-service/internal/boiler-worker/service"
	"alla/worker-service/internal/pool"
	"alla/worker-service/metrics"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
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

	if err := run(logger); err != nil {
		logger.Error("worker stopped", slog.String("err", err.Error()))
		os.Exit(1)
	}

}

func run(logger *slog.Logger) error {
	kafkaClient, err := NewKafkaClient(logger)
	if err != nil {
		return err
	}
	defer kafkaClient.Close()

	DB_SERVICE_GRPC := os.Getenv("DB_SERVICE_GRPC")
	conn, err := grpc.NewClient(DB_SERVICE_GRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to connect to service by gRPC",
			slog.String("err", err.Error()),
		)
		return fmt.Errorf("connection error: %w", err)
	}
	defer conn.Close()

	brewingClient := pb.NewJobServiceClient(conn)

	repoBrewing := repository.NewRepoBrewing(brewingClient)

	serviceBrewing := service.NewBoilerWorker(repoBrewing, time.Second*10)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":9092", nil)
	}()

	size, err := poolSize(logger)
	if err != nil {
		return err
	}

	workerPool := pool.New(size)
	defer workerPool.Stop()

	metrics.PoolSize.Set(float64(size))

	logger.Info("Starting Kafka consumer")
	for {
		fetches := kafkaClient.PollFetches(ctx)
		if ctx.Err() != nil || fetches.IsClientClosed() {
			break
		}

		start := time.Now()
		var failed atomic.Bool
		var batch sync.WaitGroup
		fetches.EachRecord(func(r *kgo.Record) {
			batch.Add(1)
			workerPool.Submit(func() {
				defer batch.Done()
				err := ProcessRecord(ctx, r, kafkaClient, serviceBrewing, logger)
				if err != nil {
					failed.Store(true)
					logger.Error("process record", slog.String("err", err.Error()))
				}
			})
		})
		batch.Wait()

		metrics.BatchSize.Observe(float64(fetches.NumRecords()))
		metrics.BatchDuration.Observe(time.Since(start).Seconds())

		if ctx.Err() != nil {
			logger.Info("planned stopping")
			return nil
		}
		if failed.Load() {
			logger.Error("one of batch is failed")
			return fmt.Errorf("error finished job")
		}

		if err := kafkaClient.CommitUncommittedOffsets(ctx); err != nil {
			logger.Warn("Kafka commit",
				slog.String("err", err.Error()),
			)
		}
	}

	logger.Info("Shutting down")
	return nil
}

func poolSize(logger *slog.Logger) (int, error) {
	const defaultSize = 5

	raw := os.Getenv("WORKER_POOL_SIZE")
	if raw == "" {
		return defaultSize, nil
	}

	size, err := strconv.Atoi(raw)
	if err != nil || size <= 0 {
		return 0, fmt.Errorf("invalid worker pool size %q", raw)
	}
	return size, nil
}
