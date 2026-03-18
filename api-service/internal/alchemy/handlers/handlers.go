package handlers

import (
	"alla/api-service/internal/alchemy/service"
	dto "alla/shared/DTO"
	errorList "alla/shared/errorList"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type GuildHandler struct {
	service  service.ServiceInterface
	validate *validator.Validate
}

func NewGuildHandler(service service.ServiceInterface) *GuildHandler {
	return &GuildHandler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *GuildHandler) sendError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError

	switch {
	case errors.Is(err, errorList.ErrWrongJsonFormat) || errors.Is(err, errorList.ErrInconsistentData):
		statusCode = http.StatusBadRequest
	case errors.Is(err, errorList.ErrRecipeAlreadyExist) || (errors.Is(err, errorList.ErrIngredientAlreadyExist)):
		statusCode = http.StatusConflict
	case errors.Is(err, errorList.ErrIngredientNotFound) || errors.Is(err, errorList.ErrRecipeNotFound):
		statusCode = http.StatusNotFound
	case errors.Is(err, errorList.ErrIngredientNotEnough):
		statusCode = http.StatusUnprocessableEntity
	case errors.Is(err, errorList.ErrCreateConnectionRecipeIngredient):
		statusCode = http.StatusUnprocessableEntity
	}

	status := struct {
		Message string
		Time    time.Time
	}{
		Message: fmt.Sprintf("Api-service: %v", err),
		Time:    time.Now(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(status)
}

func (h *GuildHandler) sendResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("Failed to encode json: %v\n", err)
	}
}

// BuyNewIngredients godoc
// @Summary      Create and add new ingredients
// @Tags         ingredients
// @Accept       json
// @Produce      json
// @Param        ingredient body dto.IngredientDTO true "Ingredient"
// @Success      201 {object} dto.IngredientResponseDTO
// @Failure      409 {string} string "already exists"
// @Router       /ingredients [post]
func (h *GuildHandler) BuyNewIngredients(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	var inputIng = dto.IngredientDTO{}

	err := json.NewDecoder(r.Body).Decode(&inputIng)
	if err != nil {
		h.sendError(w, errorList.ErrWrongJsonFormat)
		return
	}

	if err := h.validate.Struct(inputIng); err != nil {
		h.sendError(w, errorList.ErrInconsistentData)
		return
	}

	m, err := h.service.PostIngredients(ctx, inputIng)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendResponse(w, http.StatusCreated, m)
}

// ShowIngredients godoc
// @Summary      Create and add new ingredients
// @Tags         ingredients
// @Produce      json
// @Success      200 {array} dto.IngredientResponseDTO
// @Failure      500 {object} object
// @Router       /ingredients [get]
func (h *GuildHandler) ShowIngredients(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	m, err := h.service.GetIngredients(ctx)
	if err != nil {
		h.sendError(w, err)
		return
	}
	h.sendResponse(w, http.StatusOK, m)
}

// CreateRecipe godoc
// @Summary      Create new recipe
// @Tags         recipes
// @Accept       json
// @Produce      json
// @Param        recipe body dto.RecipeDTO true "Recipe"
// @Success      201 {object} dto.IngredientResponseDTO
// @Failure      409 {object} object "recipe already exists"
// @Failure      400 {object} object "invalid request"
// @Router       /recipes [post]
func (h *GuildHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var inputRecipe dto.RecipeDTO

	err := json.NewDecoder(r.Body).Decode(&inputRecipe)
	if err != nil {
		h.sendError(w, errorList.ErrWrongJsonFormat)
		return
	}

	err = h.validate.Struct(inputRecipe)
	if err != nil {
		h.sendError(w, errorList.ErrInconsistentData)
		return
	}

	m, err := h.service.PostRecipe(ctx, inputRecipe)
	if err != nil {
		h.sendError(w, err)
		return
	}
	h.sendResponse(w, http.StatusCreated, m)
}

// ShowRecipes godoc
// @Summary      Get all recipes
// @Tags         recipes
// @Produce      json
// @Param        recipe body dto.IngredientDTO true "Recipe"
// @Success      201 {array} dto.IngredientResponseDTO
// @Failure      500 {object} object
// @Router       /recipes [get]
func (h *GuildHandler) ShowRecipes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	m, err := h.service.GetRecipes(ctx)
	if err != nil {
		h.sendError(w, err)
		return
	}
	h.sendResponse(w, http.StatusOK, m)
}

// BuyExistIngredient godoc
// @Summary      Add quantity to exist ingredient
// @Tags         ingredients
// @Accept       json
// @Produce      json
// @Param        recipe   path int               true "Ingredient ID"
// @Param        quantity body dto.IngredientDTO true "Quantity to add"
// @Success      200 {object} dto.IngredientAddDTO
// @Failure      400 {object} object "invalid request"
// @Failure      404 {object} object "ingredient not found"
// @Router       /ingredients/{id} [patch]
func (h *GuildHandler) BuyExistIngredient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	ingredientIDstr := vars["id"]

	ingredientID, err := strconv.Atoi(ingredientIDstr)
	if err != nil {
		h.sendError(w, errorList.ErrWrongJsonFormat)
		return
	}

	quantity := dto.IngredientAddDTO{}
	err = json.NewDecoder(r.Body).Decode(&quantity)
	if err != nil {
		h.sendError(w, errorList.ErrWrongJsonFormat)
		return
	}
	if err := h.validate.Struct(quantity); err != nil {
		log.Println(err)
		h.sendError(w, errorList.ErrInconsistentData)
		return
	}

	req := dto.UpdateIngredientQuantityDTO{
		ID:       ingredientID,
		Quantity: quantity.Quantity,
	}

	err = h.service.AddIngredients(ctx, req)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendResponse(w, http.StatusOK, quantity)
}
