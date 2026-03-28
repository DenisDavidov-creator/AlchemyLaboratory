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
	pb.UnimplementedRecipesServiceServer
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

func (h *GrpcAlchemyHandler) GetRecipes(ctx context.Context, empty *pb.Empty) (*pb.RecipeListResponse, error) {

	recipes, err := h.service.GetRecipes(ctx)
	if err != nil {
		return nil, grpcStatus.Error(codes.NotFound, err.Error())
	}

	var requestiRecipes = pb.RecipeListResponse{}

	for _, value := range recipes {

		var recipe = pb.RecipeResponse{
			Id:                 int32(value.ID),
			Name:               value.Name,
			Description:        value.Description,
			BrewingTimeSeconds: int32(value.BrewingTimeSeconds),
		}

		for _, valueIng := range value.Ingredients {
			ing := pb.RecipeIngredients{
				IngredietnID:    int32(valueIng.IngredientID),
				QueuntityNeeded: int32(valueIng.QuantityNeeded),
			}
			recipe.RecIngs = append(recipe.RecIngs, &ing)
		}

		requestiRecipes.Recipes = append(requestiRecipes.Recipes, &recipe)
	}

	return &requestiRecipes, nil
}

func (h *GrpcAlchemyHandler) CreateRecipe(ctx context.Context, req *pb.PostRecipeRequest) (*pb.RecipeResponse, error) {

	var recipeReq = dto.RecipeDTO{
		Name:               req.Name,
		Description:        req.Description,
		BrewingTimeSeconds: int(req.BrewingTimeSeconds),
	}

	for _, valueIng := range req.RecIngs {
		ing := dto.RecipeIngredientsDTO{
			IngredientID:   int(valueIng.IngredietnID),
			QuantityNeeded: int(valueIng.QueuntityNeeded),
		}
		recipeReq.Ingredients = append(recipeReq.Ingredients, ing)
	}

	recipe, err := h.service.PostRecipe(ctx, recipeReq)
	if err != nil {
		if errors.Is(err, errorList.ErrRecipeAlreadyExist) {
			return nil, grpcStatus.Error(codes.AlreadyExists, err.Error())
		}
		return nil, grpcStatus.Error(codes.Internal, err.Error())
	}

	var requestiRecipes = pb.RecipeResponse{
		Id:                 int32(recipe.ID),
		Name:               recipe.Name,
		Description:        recipe.Description,
		BrewingTimeSeconds: int32(recipe.BrewingTimeSeconds),
	}

	for _, valueIng := range recipe.Ingredients {
		ing := pb.RecipeIngredients{
			IngredietnID:    int32(valueIng.IngredientID),
			QueuntityNeeded: int32(valueIng.QuantityNeeded),
		}
		requestiRecipes.RecIngs = append(requestiRecipes.RecIngs, &ing)
	}

	return &requestiRecipes, nil
}
