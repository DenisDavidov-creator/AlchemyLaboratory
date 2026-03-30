package repository

import (
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	"alla/shared/pb"
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RepositoryInterface interface {
	PostIngredients(context.Context, dto.IngredientDTO) (*dto.IngredientResponseDTO, error)
	GetIngredients(context.Context) ([]dto.IngredientResponseDTO, error)
	PostRecipe(context.Context, dto.RecipeDTO) (*dto.RecipeResponseDTO, error)
	GetRecipes(context.Context) ([]dto.RecipeResponseDTO, error)
	AddIngredients(context.Context, dto.UpdateIngredientQuantityDTO) error
}

type RepositoryAPI struct {
	ingredientClient pb.IngredientServiceClient
	recipeClient     pb.RecipesServiceClient
}

func NewRepository(ingredientClient pb.IngredientServiceClient, recipeClient pb.RecipesServiceClient) *RepositoryAPI {
	return &RepositoryAPI{
		ingredientClient: ingredientClient,
		recipeClient:     recipeClient,
	}
}

func (r *RepositoryAPI) PostIngredients(ctx context.Context, req dto.IngredientDTO) (*dto.IngredientResponseDTO, error) {

	resp, err := r.ingredientClient.CreateIngredient(ctx, &pb.CreateIngredientRequest{
		Name:        req.Name,
		Description: req.Description,
		Quantity:    int32(req.Quantity),
	})
	//TODO finish error handle

	if status.Code(err) == codes.AlreadyExists {
		return nil, fmt.Errorf("PostIngredients: %w", err)
	}
	return &dto.IngredientResponseDTO{
		ID:          int(resp.Id),
		Name:        resp.Name,
		Description: resp.Description,
		Quantity:    int(resp.Quantity),
	}, nil
}
func (r *RepositoryAPI) GetIngredients(ctx context.Context) ([]dto.IngredientResponseDTO, error) {

	resp, err := r.ingredientClient.GetIngredients(ctx, &pb.Empty{})
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("GetIngredients: unexpected statusCode %d", status.Code(err))
	}
	var ings = []dto.IngredientResponseDTO{}

	for _, value := range resp.Ingredietns {
		ing := dto.IngredientResponseDTO{
			ID:          int(value.Id),
			Name:        value.Name,
			Description: value.Description,
			Quantity:    int(value.Quantity),
		}
		ings = append(ings, ing)
	}
	return ings, nil

}
func (r *RepositoryAPI) PostRecipe(ctx context.Context, req dto.RecipeDTO) (*dto.RecipeResponseDTO, error) {

	var reqRecipe = pb.PostRecipeRequest{
		Name:               req.Name,
		Description:        req.Description,
		BrewingTimeSeconds: int32(req.BrewingTimeSeconds),
	}
	for _, value := range req.Ingredients {
		reqRecipe.RecIngs = append(reqRecipe.RecIngs, &pb.RecipeIngredients{
			IngredietnID:    int32(value.IngredientID),
			QueuntityNeeded: int32(value.QuantityNeeded),
		})
	}

	resp, err := r.recipeClient.CreateRecipe(ctx, &reqRecipe)

	if err != nil {
		return nil, fmt.Errorf("GetRcipes: %w", err)
	}

	var ings = []dto.RecipeIngredientsDTO{}
	for _, valueIng := range resp.RecIngs {
		ing := dto.RecipeIngredientsDTO{
			IngredientID:   int(valueIng.IngredietnID),
			QuantityNeeded: int(valueIng.QueuntityNeeded),
		}
		ings = append(ings, ing)
	}

	recipe := dto.RecipeResponseDTO{
		ID:                 int(resp.Id),
		Name:               resp.Name,
		Description:        resp.Description,
		BrewingTimeSeconds: int(resp.BrewingTimeSeconds),
		Ingredients:        ings,
	}

	return &recipe, err
}

func (r *RepositoryAPI) GetRecipes(ctx context.Context) ([]dto.RecipeResponseDTO, error) {

	resp, err := r.recipeClient.GetRecipes(ctx, &pb.Empty{})
	if err != nil {
		log.Println(err)
		if status.Code(err) == codes.Internal {
			return nil, errorList.ErrInconsistentData
		}
	}
	var results []dto.RecipeResponseDTO
	for _, value := range resp.Recipes {

		var ings = []dto.RecipeIngredientsDTO{}
		for _, valueIng := range value.RecIngs {
			ing := dto.RecipeIngredientsDTO{
				IngredientID:   int(valueIng.IngredietnID),
				QuantityNeeded: int(valueIng.QueuntityNeeded),
			}
			ings = append(ings, ing)
		}

		recipe := dto.RecipeResponseDTO{
			ID:                 int(value.Id),
			Name:               value.Name,
			Description:        value.Description,
			BrewingTimeSeconds: int(value.BrewingTimeSeconds),
			Ingredients:        ings,
		}
		results = append(results, recipe)
	}
	return results, nil

}

func (r *RepositoryAPI) AddIngredients(ctx context.Context, req dto.UpdateIngredientQuantityDTO) error {

	_, err := r.ingredientClient.AddIngredient(ctx, &pb.AddIngredientRequest{
		Id:       int32(req.ID),
		Quantity: int32(req.Quantity),
	})

	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			return errorList.ErrIngredientNotFound
		default:
			return fmt.Errorf("AddIngredietns: unexpected StatusCode %d", status.Code(err))
		}
	}

	return nil
}
