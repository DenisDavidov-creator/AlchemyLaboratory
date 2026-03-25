package handler

import (
	"alla/db-service/internal/alchemy/service"
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	"alla/shared/pb"
	"context"
	"errors"

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
		return nil, err
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
