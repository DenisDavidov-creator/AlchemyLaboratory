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
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	logger := NewLogger()
	db := connectToDB(logger)
	defer db.Close()

	redeisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
	err := redeisClient.Ping(context.Background()).Err()
	if err != nil {
		logger.Error("Failed start redis", slog.String("error", err.Error()))
		os.Exit(1)
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
		logger.Error("Failed listen gRPC", slog.String("error", err.Error()))
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingUnaryInterceptors(logger)),
	)
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	pb.RegisterIngredientServiceServer(grpcServer, grpcAlchemyHandler)
	pb.RegisterRecipesServiceServer(grpcServer, grpcAlchemyHandler)
	pb.RegisterJobServiceServer(grpcServer, grpcJobHandler)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-quit
		logger.Info("shouting down")
		grpcServer.GracefulStop()
		redeisClient.Close()
		db.Close()
	}()

	logger.Info("Start gRPC")

	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("gRPC not Running", slog.String("Error", err.Error()))
		os.Exit(1)
	}

}
