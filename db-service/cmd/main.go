package main

import (
	alchemyHandler "alla/db-service/internal/alchemy/handler"
	alchemyRepository "alla/db-service/internal/alchemy/repository"
	alchemyService "alla/db-service/internal/alchemy/service"
	brewingHandler "alla/db-service/internal/brewing/handler"
	brewingRepository "alla/db-service/internal/brewing/repository"
	brewingService "alla/db-service/internal/brewing/service"
	"alla/db-service/internal/transactor"
	"alla/shared/pb"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/subosito/gotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	db := connectToDB()
	defer db.Close()

	redeisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
	err := redeisClient.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalf("Error Start redis, %v", err)
	}
	tm := transactor.NewPtransactor(db)

	AlchemyRepository := alchemyRepository.NewAlchemyRepository(db)
	BrewingRepository := brewingRepository.NewBrewingRepository(db)

	BrewingService := brewingService.NewBrewingService(BrewingRepository, AlchemyRepository, tm)
	AlchemyService := alchemyService.NewAlchemyService(AlchemyRepository, tm, redeisClient)

	grpcAlchemyHandler := alchemyHandler.NeWGrpcAlchemicalHandler(AlchemyService)
	grpcJobHandler := brewingHandler.NeWGrpcBrewingHandler(BrewingService)

	lis, err := net.Listen("tcp", os.Getenv("DB_SERVICE_GRPC_PORT"))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	pb.RegisterIngredientServiceServer(grpcServer, grpcAlchemyHandler)
	pb.RegisterRecipesServiceServer(grpcServer, grpcAlchemyHandler)
	pb.RegisterJobServiceServer(grpcServer, grpcJobHandler)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-quit
		log.Println("shutting down...")

		grpcServer.GracefulStop()
		redeisClient.Close()
		db.Close()
	}()

	log.Println("start gRPC")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC not running: %v", err)
	}

}

func connectToDB() *sqlx.DB {
	err := gotenv.Load()
	if err != nil {
		log.Fatal("We can't get .env parameterth", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dns := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sqlx.Open("postgres", dns)

	log.Println("Connect to DB")
	if err != nil {
		log.Fatalf("We can't connect to DB: %v", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error Connect Ping: %v", err)
	}
	return db
}
