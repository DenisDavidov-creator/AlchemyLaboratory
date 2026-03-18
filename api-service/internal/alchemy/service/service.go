package service

import (
	"alla/api-service/internal/alchemy/repository"
	dto "alla/shared/DTO"
	"context"
	"fmt"
)

//go:generate mockery --name=ServiceInterface
type ServiceInterface interface {
	PostIngredients(context.Context, dto.IngredientDTO) (*dto.IngredientResponseDTO, error)
	GetIngredients(context.Context) ([]dto.IngredientResponseDTO, error)
	PostRecipe(context.Context, dto.RecipeDTO) (*dto.RecipeResponseDTO, error)
	GetRecipes(context.Context) ([]dto.RecipeResponseDTO, error)
	AddIngredients(context.Context, dto.UpdateIngredientQuantityDTO) error
}

type ServiceAPI struct {
	repo repository.RepositoryInterface
}

func NewServiceAPI(repo repository.RepositoryInterface) *ServiceAPI {
	return &ServiceAPI{
		repo: repo,
	}
}
func (s *ServiceAPI) PostIngredients(ctx context.Context, ingDTO dto.IngredientDTO) (*dto.IngredientResponseDTO, error) {
	ing, err := s.repo.PostIngredients(ctx, ingDTO)

	if err != nil {
		return nil, fmt.Errorf("Service: %w", err)
	}

	return ing, nil
}

func (s *ServiceAPI) GetIngredients(ctx context.Context) ([]dto.IngredientResponseDTO, error) {
	ings, err := s.repo.GetIngredients(ctx)
	if err != nil {
		return nil, err
	}
	return ings, nil
}

func (s *ServiceAPI) PostRecipe(ctx context.Context, recipeDTO dto.RecipeDTO) (*dto.RecipeResponseDTO, error) {

	ingID, err := s.repo.PostRecipe(ctx, recipeDTO)
	if err != nil {
		return nil, err
	}
	return ingID, nil

}

func (s *ServiceAPI) GetRecipes(ctx context.Context) ([]dto.RecipeResponseDTO, error) {

	recipe, err := s.repo.GetRecipes(ctx)
	if err != nil {
		return nil, err
	}
	return recipe, nil

}

func (s *ServiceAPI) AddIngredients(ctx context.Context, req dto.UpdateIngredientQuantityDTO) error {
	err := s.repo.AddIngredients(ctx, req)
	if err != nil {
		return err
	}
	return nil
}
