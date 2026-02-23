package handlers

import (
	dto "alchemicallabaratory/DTO"
	"alchemicallabaratory/errorList"
	"alchemicallabaratory/models"
	"alchemicallabaratory/repository"
	"alchemicallabaratory/workers/boiler"
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
	repo     repository.GrimoireRepoInterface
	boil     boiler.BoilerWorkerInterface
	validate *validator.Validate
}

func NewGuildHandler(repo repository.GrimoireRepoInterface, boil boiler.BoilerWorkerInterface) *GuildHandler {
	return &GuildHandler{
		repo:     repo,
		boil:     boil,
		validate: validator.New(),
	}
}

func (h *GuildHandler) sendError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError

	switch {
	case errors.Is(err, errorList.ErrWrongJsonFormat) || errors.Is(err, errorList.ErrInconsistencyData):
		statusCode = http.StatusBadRequest
	}

	status := struct {
		Message string
		Time    time.Time
	}{
		Message: fmt.Sprintf("We get error: %v", err),
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

func (h *GuildHandler) BuyNewIngredients(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	var inputIngredients = dto.IngredientDTO{}

	err := json.NewDecoder(r.Body).Decode(&inputIngredients)
	if err != nil {
		h.sendError(w, errorList.ErrWrongJsonFormat)
		return
	}

	if err := h.validate.Struct(inputIngredients); err != nil {
		h.sendError(w, errorList.ErrInconsistencyData)
		return
	}

	m, err := h.repo.PostIngredients(ctx, models.IngredientsToModel(inputIngredients))
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendResponse(w, http.StatusCreated, m)
}

func (h *GuildHandler) ShowIngredients(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	m, err := h.repo.GetIngredients(ctx)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendResponse(w, http.StatusOK, m)

}

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
		h.sendError(w, errorList.ErrInconsistencyData)
		return
	}

	m, err := h.repo.PostRecipe(ctx, models.RecipeToMdel(inputRecipe))
	if err != nil {
		h.sendError(w, err)
		return
	}
	h.sendResponse(w, http.StatusCreated, m)
}

func (h *GuildHandler) ShowRecipes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	m, err := h.repo.GetRecipes(ctx)
	if err != nil {
		h.sendError(w, err)
		return
	}
	h.sendResponse(w, http.StatusOK, m)
}

func (h *GuildHandler) Brew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var recipeID dto.JobDTO

	err := json.NewDecoder(r.Body).Decode(&recipeID)
	if err != nil {
		h.sendError(w, errorList.ErrWrongJsonFormat)
		return
	}

	if err := h.validate.Struct(recipeID); err != nil {
		h.sendError(w, errorList.ErrInconsistencyData)
		return
	}

	m, err := h.repo.PostJob(ctx, models.BrewingJobsToModel(recipeID))
	if err != nil {
		h.sendError(w, err)
		return
	}

	err = h.boil.Boiled(ctx, m.PublicID)
	if err != nil {
		h.sendError(w, err)
		return
	}
	h.sendResponse(w, http.StatusOK, "Your order is ready")
}

func (h *GuildHandler) StatusBrew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uuid := r.URL.Query().Get("uuid")
	status, err := h.repo.GetBrewStatus(ctx, uuid)
	if err != nil {
		h.sendError(w, err)
		return
	}
	h.sendResponse(w, http.StatusOK, status)
}

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
		h.sendError(w, errorList.ErrInconsistencyData)
		return
	}

	err = h.repo.AddIngredients(ctx, ingredientID, quantity.Quantity)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendResponse(w, http.StatusOK, quantity)
}
