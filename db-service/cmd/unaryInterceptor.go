package main

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"

	"alla/db-service/metrics"
)

func loggingUnaryInterceptors(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}
		logger.Info("gRPC request", slog.String("Method", info.FullMethod))

		resp, err := handler(ctx, req)

		metrics.GrpcRequestsDuration.WithLabelValues(info.FullMethod).Observe(time.Since(start).Seconds())
		if err != nil {
			logger.Error("gRPC error",
				slog.String("Method", info.FullMethod),
				slog.String("error", err.Error()),
			)
			metrics.GrpcRequestsTotal.WithLabelValues(info.FullMethod, "error").Inc()

			return nil, err
		}
		metrics.GrpcRequestsTotal.WithLabelValues(info.FullMethod, "success").Inc()

		return resp, err
	}
}
