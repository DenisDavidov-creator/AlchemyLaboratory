package main

import (
	alchemyHandler "alla/api-service/internal/alchemy/handlers"
	alchemyRepo "alla/api-service/internal/alchemy/repository"
	alchemyService "alla/api-service/internal/alchemy/service"
	brewingHandler "alla/api-service/internal/brewing/handlers"
	brewingRepo "alla/api-service/internal/brewing/repository"
	brewingService "alla/api-service/internal/brewing/service"
	"alla/api-service/internal/server"
	"log/slog"

	pb "alla/shared/pb"

	"os"

	"github.com/subosito/gotenv"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @title           AlchemicalLab
// @version         2.0
// @description     Website for brewing potions and elixirs.
// @host            localhost:8080
// @BasePath        /
func main() {

	logger := NewLogger()

	err := gotenv.Load()
	if err != nil {
		logger.Error("get .env", slog.String("error", err.Error()))
		os.Exit(1)
	}

	DB_SERVICE_GRPC := os.Getenv("DB_SERVICE_GRPC")
	WORKER_SERVICE_GRPC := os.Getenv("WORKER_SERVICE_GRPC")

	connDB, err := grpc.NewClient(DB_SERVICE_GRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Connect to DB-service by gRPC", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer connDB.Close()

	connWorker, err := grpc.NewClient(WORKER_SERVICE_GRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Connect to worker-service by gRPC", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer connWorker.Close()

	kafkaClient, err := kgo.NewClient(
		kgo.SeedBrokers(os.Getenv("KAFKA_ADDR")),
	)
	if err != nil {
		logger.Error("Connect to Kafka", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer kafkaClient.Close()

	ingredientClient := pb.NewIngredientServiceClient(connDB)
	recipeClient := pb.NewRecipesServiceClient(connDB)
	jobClient := pb.NewJobServiceClient(connDB)

	repoA := alchemyRepo.NewRepository(ingredientClient, recipeClient)
	repoBrewing := brewingRepo.NewBrewingRepo(jobClient, kafkaClient)

	serviceA := alchemyService.NewServiceAPI(repoA)
	serviceBrewing := brewingService.NewBrewingService(repoBrewing)

	handlerA := alchemyHandler.NewGuildHandler(serviceA)
	handlerBrewing := brewingHandler.NewBrewingHandler(serviceBrewing)

	serverAPI := server.NewServer(handlerA, handlerBrewing, logger)

	err = serverAPI.Run()
	if err != nil {
		logger.Error("Start server", slog.String("error", err.Error()))
		os.Exit(1)
	}

}
