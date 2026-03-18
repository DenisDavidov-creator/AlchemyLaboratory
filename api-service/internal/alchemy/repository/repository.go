package repository

import (
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type RepositoryInterface interface {
	PostIngredients(context.Context, dto.IngredientDTO) (*dto.IngredientResponseDTO, error)
	GetIngredients(context.Context) ([]dto.IngredientResponseDTO, error)
	PostRecipe(context.Context, dto.RecipeDTO) (*dto.RecipeResponseDTO, error)
	GetRecipes(context.Context) ([]dto.RecipeResponseDTO, error)
	AddIngredients(context.Context, dto.UpdateIngredientQuantityDTO) error
}

type RepositoryAPI struct {
	baseURL    string
	httpClient *http.Client
}

func NewRepository(URL string) *RepositoryAPI {
	return &RepositoryAPI{
		baseURL:    URL,
		httpClient: &http.Client{},
	}
}

func (r *RepositoryAPI) PostIngredients(ctx context.Context, req dto.IngredientDTO) (*dto.IngredientResponseDTO, error) {

	body, err := json.Marshal(req)
	if err != nil {
		return nil, errorList.ErrWrongJsonFormat
	}

	url := fmt.Sprintf("%s/internal/ingredients", r.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))

	httpReq.Header.Set("Content-type", "application/json")

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusConflict:
		return nil, errorList.ErrIngredientAlreadyExist
	case http.StatusCreated:
		var result dto.IngredientResponseDTO
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return nil, errorList.ErrWrongJsonFormat
		}
		return &result, nil
	default:
		return nil, fmt.Errorf("PostIngredients: unexpected StatusCode: %d", resp.StatusCode)
	}
}
func (r *RepositoryAPI) GetIngredients(ctx context.Context) ([]dto.IngredientResponseDTO, error) {

	url := fmt.Sprintf("%s/internal/ingredients", r.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	httpReq.Header.Set("Content-type", "application/json")

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var result []dto.IngredientResponseDTO
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return nil, errorList.ErrWrongJsonFormat
		}
		return result, nil
	default:
		return nil, fmt.Errorf("GetIngredients: unexpected StatusCode: %d", resp.StatusCode)
	}

}
func (r *RepositoryAPI) PostRecipe(ctx context.Context, req dto.RecipeDTO) (*dto.RecipeResponseDTO, error) {

	body, err := json.Marshal(req)
	if err != nil {
		return nil, errorList.ErrWrongJsonFormat
	}

	url := fmt.Sprintf("%s/internal/recipes", r.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))

	httpReq.Header.Set("Content-type", "application/json")

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated:
		var result dto.RecipeResponseDTO
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return nil, err
		}
		return &result, nil
	case http.StatusConflict:
		return nil, errorList.ErrRecipeAlreadyExist
	default:
		return nil, fmt.Errorf("PostRecipe: unexpected StatusCode: %d", resp.StatusCode)
	}
}

func (r *RepositoryAPI) GetRecipes(ctx context.Context) ([]dto.RecipeResponseDTO, error) {

	url := fmt.Sprintf("%s/internal/recipes", r.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	httpReq.Header.Set("Content-type", "application/json")

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var result []dto.RecipeResponseDTO
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return nil, err
		}
		return result, nil
	case http.StatusBadRequest:
		return nil, errorList.ErrInconsistentData
	default:
		return nil, fmt.Errorf("GetRecipes: unexpected StatusCode: %d", resp.StatusCode)
	}

}

func (r *RepositoryAPI) AddIngredients(ctx context.Context, req dto.UpdateIngredientQuantityDTO) error {

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("Repository: %w", err)
	}

	url := fmt.Sprintf("%s/internal/ingredients/%d", r.baseURL, req.ID)
	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(body))

	httpReq.Header.Set("Content-type", "application/json")

	resp, err := r.httpClient.Do(httpReq)

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return errorList.ErrIngredientNotFound
	default:
		return fmt.Errorf("GetRecipes: unexpected StatusCode: %d", resp.StatusCode)
	}

}
