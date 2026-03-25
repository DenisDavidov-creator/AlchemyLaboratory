package service

import (
	"alla/db-service/internal/alchemy/repository"
	"alla/db-service/internal/transactor"
	"alla/db-service/models"
	dto "alla/shared/DTO"
	errorList "alla/shared/errorList"
	"context"
	"fmt"
)

//go:generate mockery --name=AlchemyServiceInterface
type AlchemyServiceInterface interface {
	PostIngredients(context.Context, dto.IngredientDTO) (*dto.IngredientResponseDTO, error)
	GetIngredients(context.Context) ([]dto.IngredientResponseDTO, error)
	PostRecipe(context.Context, dto.RecipeDTO) (*dto.RecipeResponseDTO, error)
	GetRecipes(context.Context) ([]dto.RecipeResponseDTO, error)
	AddIngredients(ctx context.Context, ingID int, quantity int) error
}

type AclhemyService struct {
	aRepo repository.AlchemyRepoInterface
	trans transactor.TransactorInterface
}

func NewAlchemyService(aRepo repository.AlchemyRepoInterface, trans transactor.TransactorInterface) *AclhemyService {
	return &AclhemyService{
		aRepo: aRepo,
		trans: trans,
	}
}

func (s *AclhemyService) PostIngredients(ctx context.Context, ingDTO dto.IngredientDTO) (*dto.IngredientResponseDTO, error) {

	exists, err := s.aRepo.CheckIngredientExistsByName(ctx, ingDTO.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("PostIngredient: %w", errorList.ErrIngredientAlreadyExist)
	}

	newIng := models.Ingredient{
		Name:        ingDTO.Name,
		Description: ingDTO.Description,
		Quantity:    ingDTO.Quantity,
	}

	ing, err := s.aRepo.PostIngredients(ctx, newIng)

	if err != nil {
		return nil, fmt.Errorf("PostIngredient: %w", err)
	}

	responseIngredient := dto.IngredientResponseDTO{
		ID:          ing.ID,
		Name:        ing.Name,
		Description: ing.Description,
		Quantity:    ing.Quantity,
	}

	return &responseIngredient, nil

}

func (s *AclhemyService) GetIngredients(ctx context.Context) ([]dto.IngredientResponseDTO, error) {

	ings, err := s.aRepo.GetIngredients(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetIngredients: %w", err)
	}
	var ingsResponse []dto.IngredientResponseDTO
	for _, value := range ings {
		ing := dto.IngredientResponseDTO{
			ID:          value.ID,
			Name:        value.Name,
			Description: value.Description,
			Quantity:    value.Quantity,
		}
		ingsResponse = append(ingsResponse, ing)
	}

	return ingsResponse, nil

}

func (s *AclhemyService) PostRecipe(ctx context.Context, recipeDTO dto.RecipeDTO) (*dto.RecipeResponseDTO, error) {
	recipeModels := models.Recipe{
		Name:               recipeDTO.Name,
		Description:        recipeDTO.Description,
		BrewingTimeSeconds: recipeDTO.BrewingTimeSeconds,
	}

	exist, err := s.aRepo.CheckExistRecipeByName(ctx, recipeDTO.Name)

	if err != nil {
		return nil, fmt.Errorf("PostRecipe: %w", err)
	}
	if exist {
		return nil, errorList.ErrRecipeAlreadyExist
	}
	err = s.trans.WithinTransaction(ctx, func(ctx context.Context) error {
		err = s.aRepo.CreateRecipe(ctx, &recipeModels)
		if err != nil {
			return fmt.Errorf("PostRecipe: %w", err)
		}

		var valueStrings = []string{}
		var valueArgs = []any{}

		for i, value := range recipeDTO.Ingredients {
			p1 := i*3 + 1
			p2 := i*3 + 2
			p3 := i*3 + 3
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", p1, p2, p3))
			valueArgs = append(valueArgs, recipeModels.ID, value.IngredientID, value.QuantityNeeded)

			recipeModels.Ingredients = append(recipeModels.Ingredients, models.RecipeIngredients{
				RecipeID:       recipeModels.ID,
				IngredientID:   value.IngredientID,
				QuantityNeeded: value.QuantityNeeded,
			})

		}

		err = s.aRepo.CreateRecipeIngredients(ctx, valueStrings, valueArgs)
		if err != nil {
			return fmt.Errorf("PostRecipe: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	returnRecipe := dto.RecipeResponseDTO{
		ID:                 recipeModels.ID,
		Name:               recipeDTO.Name,
		Description:        recipeDTO.Description,
		BrewingTimeSeconds: recipeModels.BrewingTimeSeconds,
		Ingredients:        recipeDTO.Ingredients,
	}
	return &returnRecipe, nil
}

func (s *AclhemyService) GetRecipes(ctx context.Context) ([]dto.RecipeResponseDTO, error) {

	recipes, err := s.aRepo.GetRecipes(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetRecipes: %w", err)
	}

	var recipesResponse []dto.RecipeResponseDTO

	for _, value := range recipes {

		var resIngs []dto.RecipeIngredientsDTO
		for _, valueIng := range value.Ingredients {
			resIng := dto.RecipeIngredientsDTO{
				IngredientID:   valueIng.IngredientID,
				QuantityNeeded: valueIng.QuantityNeeded,
			}
			resIngs = append(resIngs, resIng)
		}

		recRes := dto.RecipeResponseDTO{
			ID:                 value.ID,
			Name:               value.Name,
			Description:        value.Description,
			BrewingTimeSeconds: value.BrewingTimeSeconds,
			Ingredients:        resIngs,
		}

		recipesResponse = append(recipesResponse, recRes)
	}

	return recipesResponse, nil

}

func (s *AclhemyService) AddIngredients(ctx context.Context, ingID int, quantity int) error {

	err := s.aRepo.AddIngredients(ctx, ingID, quantity)
	if err != nil {
		return fmt.Errorf("AddIngredients: %w", err)
	}

	return nil
}
