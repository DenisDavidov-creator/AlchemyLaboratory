package main

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
)

func loggingUnaryInterceptors(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}
		logger.Info("gRPC request", slog.String("Method", info.FullMethod))

		resp, err := handler(ctx, req)

		if err != nil {
			logger.Error("gRPC error",
				slog.String("Method", info.FullMethod),
				slog.String("error", err.Error()),
			)
		}

		return resp, err
	}
}
