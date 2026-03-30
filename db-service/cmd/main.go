package main

import (
	alchemyHandler "alla/db-service/internal/alchemy/handler"
	alchemyRepository "alla/db-service/internal/alchemy/repository"
	alchemyService "alla/db-service/internal/alchemy/service"
	brewingHandler "alla/db-service/internal/brewing/handler"
	brewingRepository "alla/db-service/internal/brewing/repository"
	brewingService "alla/db-service/internal/brewing/service"
	"alla/db-service/internal/server"
	"alla/db-service/internal/transactor"
	"alla/shared/pb"
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/subosito/gotenv"
	"google.golang.org/grpc"
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

	BrewingHandler := brewingHandler.NewBrewingHandler(BrewingService)
	AlchemyHandler := alchemyHandler.NewAlchemyHandler(AlchemyService)

	grpcAlchemyHandler := alchemyHandler.NeWGrpcAlchemicalHandler(AlchemyService)
	grpcJobHandler := brewingHandler.NeWGrpcBrewingHandler(BrewingService)

	//TODO add .env
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterIngredientServiceServer(grpcServer, grpcAlchemyHandler)
	pb.RegisterRecipesServiceServer(grpcServer, grpcAlchemyHandler)
	pb.RegisterJobServiceServer(grpcServer, grpcJobHandler)

	go func() {
		log.Println("start gRPC")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC not running: %v", err)
		}
	}()

	DBServer := server.NewServer(AlchemyHandler, BrewingHandler)

	err = DBServer.Run()
	if err != nil {
		log.Fatal("Error: ", err)
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
