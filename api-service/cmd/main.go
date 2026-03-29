package main

import (
	alchemyHandler "alla/api-service/internal/alchemy/handlers"
	alchemyRepo "alla/api-service/internal/alchemy/repository"
	alchemyService "alla/api-service/internal/alchemy/service"
	brewingHandler "alla/api-service/internal/brewing/handlers"
	brewingRepo "alla/api-service/internal/brewing/repository"
	brewingService "alla/api-service/internal/brewing/service"
	"alla/api-service/internal/server"

	pb "alla/shared/pb"

	"log"
	"os"

	"github.com/subosito/gotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @title           AlchemicalLab
// @version         2.0
// @description     Website for brewing potions and elixirs.
// @host            localhost:8080
// @BasePath        /
func main() {

	err := gotenv.Load()
	if err != nil {
		log.Fatal("We can't get .env parameterth", err)
	}

	DB_SERVICE_URL := os.Getenv("DB_SERVICE_URL")
	DB_SERVICE_GRPC := os.Getenv("DB_SERVICE_GRPC")
	WORKER_SERVICE_URL := os.Getenv("WORKER_SERVICE_URL")
	WORKER_SERVICE_GRPC := os.Getenv("WORKER_SERVICE_GRPC")

	connDB, err := grpc.NewClient(DB_SERVICE_GRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to service by gRPC %v", err)
	}
	defer connDB.Close()

	connWorker, err := grpc.NewClient(WORKER_SERVICE_GRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to service by gRPC %v", err)
	}
	defer connWorker.Close()

	ingredientClient := pb.NewIngredientServiceClient(connDB)
	recipeClient := pb.NewRecipesServiceClient(connDB)
	jobClient := pb.NewJobServiceClient(connDB)
	brewingClient := pb.NewBrewServiceClient(connWorker)

	repoA := alchemyRepo.NewRepository(DB_SERVICE_URL, ingredientClient, recipeClient)
	repoBrewing := brewingRepo.NewBrewingRepo(DB_SERVICE_URL, WORKER_SERVICE_URL, jobClient, brewingClient)

	serviceA := alchemyService.NewServiceAPI(repoA)
	serviceBrewing := brewingService.NewBrewingService(repoBrewing)

	handlerA := alchemyHandler.NewGuildHandler(serviceA)
	handlerBrewing := brewingHandler.NewBrewingHandler(serviceBrewing)

	serverAPI := server.NewServer(handlerA, handlerBrewing)

	err = serverAPI.Run()
	if err != nil {
		log.Fatalf("Server error")
	}

}
