package main

import (
	"alla/worker-service/internal/boiler-worker/repository"
	"alla/worker-service/internal/boiler-worker/service"
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	dto "alla/shared/DTO"
	pb "alla/shared/pb"

	"github.com/subosito/gotenv"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	err := gotenv.Load()
	if err != nil {
		log.Fatal("We can't get .env parameterth", err)
	}

	kafkaClient, err := kgo.NewClient(
		kgo.SeedBrokers(os.Getenv("KAFKA_ADDR")),
		kgo.ConsumerGroup("worker-group"),
		kgo.ConsumeTopics("brew-jobs"),
	)
	if err != nil {
		log.Fatalf("failed to connect to Kafka: %v", err)
	}
	defer kafkaClient.Close()

	DB_SERVICE_GRPC := os.Getenv("DB_SERVICE_GRPC")

	conn, err := grpc.NewClient(DB_SERVICE_GRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to service by gRPC %v", err)
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
		log.Println("Shouting down...")
		cancel()
		kafkaClient.Close()
		conn.Close()
	}()

	log.Println("Starting Kafka consumer...")
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
				log.Printf("consumer: unmarshal error: %v", err)
				return
			}
			if err := serviceBrewing.Boiled(context.Background(), dto.JobUUIDDTO{JobUUID: msg.JobUUID}); err != nil {
				log.Printf("consumer: Boiled error: %v", err)
			}
		})
	}

}
