package main

import (
	"alla/worker-service/internal/boiler-worker/handler"
	"alla/worker-service/internal/boiler-worker/repository"
	"alla/worker-service/internal/boiler-worker/service"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "alla/shared/pb"

	"github.com/subosito/gotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	err := gotenv.Load()
	if err != nil {
		log.Fatal("We can't get .env parameterth", err)
	}

	DB_SERVICE_URL := os.Getenv("DB_SERVICE_URL")
	DB_SERVICE_GRPC := os.Getenv("DB_SERVICE_GRPC")

	conn, err := grpc.NewClient(DB_SERVICE_GRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to service by gRPC %v", err)
	}
	defer conn.Close()

	brewingClient := pb.NewJobServiceClient(conn)

	repoBrewing := repository.NewRepoBrewing(DB_SERVICE_URL, brewingClient)

	serviceBrewing := service.NewBoilerWorker(repoBrewing)

	grpcBrewingHandler := handler.NewGrpcBrewingHandler(serviceBrewing)

	lis, err := net.Listen("tcp", os.Getenv("WORKER_SERVICE_GRPC_PORT"))
	if err != nil {
		log.Fatalf("failed to listen gRPC: %w", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterBrewServiceServer(grpcServer, grpcBrewingHandler)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-quit
		log.Println("Shouting down...")
		grpcServer.GracefulStop()
	}()

	log.Println("Start gRPC")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC not running: %v", err)
	}

}
