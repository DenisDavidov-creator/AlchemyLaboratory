package main

import (
	"alla/worker-service/internal/boiler-worker/repository"
	"alla/worker-service/internal/boiler-worker/service"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	logger.Info("Starting Kafka consumer")
	for {
		fetches := kafkaClient.PollFetches(ctx)
		if ctx.Err() != nil || fetches.IsClientClosed() {
			break
		}
		fetches.EachRecord(func(r *kgo.Record) {
			ProcessRecord(ctx, r, kafkaClient, serviceBrewing, logger)
		})
	}
	logger.Info("Shouting down")
}
