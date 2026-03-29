package handler

import (
	"alla/db-service/internal/brewing/service"
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	"alla/shared/pb"
	"context"
	"errors"

	"google.golang.org/grpc/codes"

	grpcStatus "google.golang.org/grpc/status"
)

type GrpcBrewingHandler struct {
	pb.UnimplementedJobServiceServer
	service service.BrewingServiceInterface
}

func NeWGrpcBrewingHandler(service service.BrewingServiceInterface) *GrpcBrewingHandler {
	return &GrpcBrewingHandler{
		service: service,
	}
}

func (h *GrpcBrewingHandler) PostJob(ctx context.Context, req *pb.PostJobRequest) (*pb.JobUUID, error) {

	var inputReq = dto.JobDTO{
		RecipeID: int(req.RecipeId),
		Details:  req.Details,
	}

	resp, err := h.service.CreateJob(ctx, inputReq)
	if err != nil {
		return nil, grpcStatus.Error(codes.Internal, err.Error())
	}

	return &pb.JobUUID{JobUUID: resp.JobUUID}, nil
}

func (h *GrpcBrewingHandler) StartBrewing(ctx context.Context, req *pb.JobUUID) (*pb.JobTime, error) {

	resp, err := h.service.StartBrewing(ctx, dto.JobUUIDDTO{JobUUID: req.JobUUID})
	if err != nil {
		if errors.Is(err, errorList.ErrJobNotFound) || errors.Is(err, errorList.ErrRecipeNotFound) {
			return nil, grpcStatus.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, errorList.ErrIngredientNotEnough) {
			return nil, grpcStatus.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, grpcStatus.Error(codes.Internal, err.Error())
	}

	return &pb.JobTime{BrewingTime: int32(resp.BrweingTime)}, nil
}

func (h *GrpcBrewingHandler) ChangeStatus(ctx context.Context, req *pb.ChangeJobStatus) (*pb.Empty, error) {

	var reqInput = dto.JobStatusDTO{
		UUID:   req.Uuid,
		Status: req.Status,
	}

	err := h.service.SetStatus(ctx, reqInput)
	if err != nil {
		if errors.Is(err, errorList.ErrJobNotFound) {
			return nil, grpcStatus.Error(codes.NotFound, err.Error())
		}
		return nil, grpcStatus.Error(codes.Internal, err.Error())
	}

	return &pb.Empty{}, nil
}

func (h *GrpcBrewingHandler) GetBrewStatus(ctx context.Context, req *pb.JobUUID) (*pb.JobStatusResponse, error) {

	resp, err := h.service.GetBrewStatus(ctx, dto.JobUUIDDTO{JobUUID: req.JobUUID})
	if err != nil {
		if errors.Is(err, errorList.ErrJobNotFound) {
			return nil, grpcStatus.Error(codes.NotFound, err.Error())
		}
		return nil, grpcStatus.Error(codes.Internal, err.Error())
	}

	return &pb.JobStatusResponse{Status: resp.Status}, nil
}
