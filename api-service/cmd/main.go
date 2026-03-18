package main

import (
	alchemyHandler "alla/api-service/internal/alchemy/handlers"
	alchemyRepo "alla/api-service/internal/alchemy/repository"
	alchemyService "alla/api-service/internal/alchemy/service"
	brewingHandler "alla/api-service/internal/brewing/handlers"
	brewingRepo "alla/api-service/internal/brewing/repository"
	brewingService "alla/api-service/internal/brewing/service"
	"alla/api-service/internal/server"

	"log"
	"os"

	"github.com/subosito/gotenv"
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
	WORKER_SERVICE_URL := os.Getenv("WORKER_SERVICE_URL")

	repoA := alchemyRepo.NewRepository(DB_SERVICE_URL)
	repoBrewing := brewingRepo.NewBrewingRepo(DB_SERVICE_URL, WORKER_SERVICE_URL)

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
