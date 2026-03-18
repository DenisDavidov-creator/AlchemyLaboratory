package main

import (
	"alla/worker-service/internal/boiler-worker/handler"
	"alla/worker-service/internal/boiler-worker/repository"
	"alla/worker-service/internal/boiler-worker/service"
	"alla/worker-service/server"
	"log"
	"os"

	"github.com/subosito/gotenv"
)

func main() {

	err := gotenv.Load()
	if err != nil {
		log.Fatal("We can't get .env parameterth", err)
	}

	DB_SERVICE_URL := os.Getenv("DB_SERVICE_URL")

	repoBrewing := repository.NewRepoBrewing(DB_SERVICE_URL)

	serviceBrewing := service.NewBoilerWorker(repoBrewing)

	handlerBrewing := handler.NewHandlerBrewing(serviceBrewing)

	serverAPI := server.NewServer(*handlerBrewing)

	err = serverAPI.Run()
	if err != nil {
		log.Fatalf("Server error")
	}

}
