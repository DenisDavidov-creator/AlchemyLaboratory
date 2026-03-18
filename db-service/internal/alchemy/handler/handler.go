package handler

import (
	"alla/db-service/internal/alchemy/service"
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type AlchemyHandler struct {
	service service.AlchemyServiceInterface
}

func NewAlchemyHandler(service service.AlchemyServiceInterface) *AlchemyHandler {
	return &AlchemyHandler{
		service: service,
	}
}

func (h *AlchemyHandler) sendError(w http.ResponseWriter, err error) {
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
	case errors.Is(err, errorList.ErrCreateConnectionRecipeIngredient) || errors.Is(err, errorList.ErrJobNotFound):
		statusCode = http.StatusUnprocessableEntity
	}

	status := struct {
		Message string
	}{
		Message: fmt.Sprintf("DB-serivce: %v", err),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(status)
}

func (h *AlchemyHandler) sendResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("Failed to encode json: %v\n", err)
	}
}

func (h *AlchemyHandler) BuyNewIngredients(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	var inputIng = dto.IngredientDTO{}

	err := json.NewDecoder(r.Body).Decode(&inputIng)
	if err != nil {
		h.sendError(w, errorList.ErrWrongJsonFormat)
		return
	}

	m, err := h.service.PostIngredients(ctx, inputIng)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendResponse(w, http.StatusCreated, m)
}

func (h *AlchemyHandler) GetIngredients(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	m, err := h.service.GetIngredients(ctx)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendResponse(w, http.StatusOK, m)

}

func (h *AlchemyHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var inputRecipe dto.RecipeDTO

	err := json.NewDecoder(r.Body).Decode(&inputRecipe)
	if err != nil {
		h.sendError(w, errorList.ErrWrongJsonFormat)
		return
	}

	m, err := h.service.PostRecipe(ctx, inputRecipe)
	if err != nil {
		h.sendError(w, err)
		return
	}
	h.sendResponse(w, http.StatusCreated, m)
}

func (h *AlchemyHandler) ShowRecipes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	m, err := h.service.GetRecipes(ctx)
	if err != nil {
		h.sendError(w, err)
		return
	}
	h.sendResponse(w, http.StatusOK, m)
}

func (h *AlchemyHandler) BuyExistIngredient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	ingredientIDstr := vars["id"]

	ingredientID, err := strconv.Atoi(ingredientIDstr)
	if err != nil {
		h.sendError(w, errorList.ErrWrongJsonFormat)
		return
	}

	quantity := dto.UpdateIngredientQuantityDTO{}
	err = json.NewDecoder(r.Body).Decode(&quantity)
	if err != nil {
		h.sendError(w, errorList.ErrWrongJsonFormat)
		return
	}

	err = h.service.AddIngredients(ctx, ingredientID, quantity.Quantity)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendResponse(w, http.StatusOK, quantity)
}
