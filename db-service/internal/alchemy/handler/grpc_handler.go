package handler

import (
	"alla/db-service/internal/alchemy/service"
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	"alla/shared/pb"
	"context"
	"errors"
	"log"

	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
)

type GrpcAlchemyHandler struct {
	pb.UnimplementedIngredientServiceServer
	service service.AlchemyServiceInterface
}

func NeWGrpcAlchemicalHandler(service service.AlchemyServiceInterface) *GrpcAlchemyHandler {
	return &GrpcAlchemyHandler{
		service: service,
	}
}

func (h *GrpcAlchemyHandler) CreateIngredient(ctx context.Context, req *pb.CreateIngredientRequest) (*pb.IngredientResponse, error) {

	var inputIng = dto.IngredientDTO{
		Name:        req.Name,
		Description: req.Description,
		Quantity:    int(req.Quantity),
	}

	m, err := h.service.PostIngredients(ctx, inputIng)
	if err != nil {
		if errors.Is(err, errorList.ErrIngredientAlreadyExist) {
			return nil, grpcStatus.Error(codes.AlreadyExists, err.Error())
		}
		return nil, grpcStatus.Error(codes.Internal, err.Error())
	}

	return &pb.IngredientResponse{
		Id:          int32(m.ID),
		Name:        m.Name,
		Description: m.Description,
		Quantity:    int32(m.Quantity),
	}, nil
}

func (h *GrpcAlchemyHandler) GetIngredients(ctx context.Context, empty *pb.Empty) (*pb.IngredientListResponse, error) {

	ings, err := h.service.GetIngredients(ctx)
	if err != nil {
		return nil, grpcStatus.Error(codes.NotFound, err.Error())
	}

	var requestiIngs = pb.IngredientListResponse{}

	for _, value := range ings {
		requestiIngs.Ingredietns = append(requestiIngs.Ingredietns, &pb.IngredientResponse{
			Id:          int32(value.ID),
			Name:        value.Name,
			Description: value.Description,
			Quantity:    int32(value.Quantity),
		})
	}
	return &requestiIngs, nil
}

func (h *GrpcAlchemyHandler) AddIngredient(ctx context.Context, req *pb.AddIngredientRequest) (*pb.Empty, error) {

	err := h.service.AddIngredients(ctx, int(req.Id), int(req.Quantity))
	if err != nil {
		log.Println(err)
		if errors.Is(err, errorList.ErrIngredientNotFound) {
			return &pb.Empty{}, grpcStatus.Error(codes.NotFound, err.Error())
		}
		return &pb.Empty{}, grpcStatus.Error(codes.Internal, err.Error())

	}

	return &pb.Empty{}, nil
}
