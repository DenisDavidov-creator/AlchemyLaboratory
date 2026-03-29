package handler

import (
	dto "alla/shared/DTO"
	"alla/shared/pb"
	"alla/worker-service/internal/boiler-worker/service"
	"context"
	"log"
)

type GrpcBrewingHandler struct {
	pb.UnimplementedBrewServiceServer
	service service.ServiceInterface
}

func NewGrpcBrewingHandler(service service.ServiceInterface) *GrpcBrewingHandler {
	return &GrpcBrewingHandler{service: service}
}

func (h *GrpcBrewingHandler) Brew(ctx context.Context, req *pb.JobUUID) (*pb.Empty, error) {
	log.Println(req.JobUUID)
	h.service.Boiled(ctx, dto.JobUUIDDTO{JobUUID: req.JobUUID})

	return &pb.Empty{}, nil
}
