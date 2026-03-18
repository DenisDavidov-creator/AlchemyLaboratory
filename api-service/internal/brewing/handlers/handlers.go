package handlers

import (
	"alla/api-service/internal/brewing/service"
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

type BrewingHandler struct {
	service  service.BrewingServiceInterface
	validate *validator.Validate
}

func NewBrewingHandler(service service.BrewingServiceInterface) *BrewingHandler {
	return &BrewingHandler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *BrewingHandler) sendError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError

	switch {
	case errors.Is(err, errorList.ErrWrongJsonFormat) || errors.Is(err, errorList.ErrInconsistentData):
		statusCode = http.StatusBadRequest
	case errors.Is(err, errorList.ErrRecipeAlreadyExist) || (errors.Is(err, errorList.ErrIngredientAlreadyExist)):
		statusCode = http.StatusConflict
	case errors.Is(err, errorList.ErrIngredientNotFound) || errors.Is(err, errorList.ErrRecipeNotFound) || errors.Is(err, errorList.ErrJobNotFound):
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

func (h *BrewingHandler) sendResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Printf("Failed to encode json: %v\n", err)
	}
}

// Brew godoc
// @Summary      Started brewing a potion
// @Tags         brewing
// @Accept       json
// @Produce      json
// @Param        brew body    dto.JobDTO true "Brew"
// @Success      200 {object} dto.JobUUIDDTO
// @Failure      400 {object} object "invalid request"
// @Failure      404 {object} object "recipe not found"
// @Failure      422 {object} object "ingredient not enough"
// @Failure      500 {object} object
// @Router       /brew [post]
func (h *BrewingHandler) Brew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var recipeID dto.JobDTO

	err := json.NewDecoder(r.Body).Decode(&recipeID)
	if err != nil {
		h.sendError(w, errorList.ErrWrongJsonFormat)
		return
	}

	if err := h.validate.Struct(recipeID); err != nil {
		h.sendError(w, errorList.ErrInconsistentData)
		return
	}

	m, err := h.service.PostJob(ctx, recipeID)
	if err != nil {
		h.sendError(w, err)
		return
	}

	h.sendResponse(w, http.StatusOK, m)
}

// Brew godoc
// @Summary      Started to brewing potion
// @Tags         brewing
// @Produce      json
// @Param		 uuid query string true "Job UUID"
// @Success      200 {object} dto.JobStatusDTO
// @Failure      404 {object} object "job not found"
// @Failure      500 {object} object
// @Router       /brew/status [get]
func (h *BrewingHandler) StatusBrew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uuid := r.URL.Query().Get("uuid")

	JobUUIDDTO := dto.JobUUIDDTO{
		JobUUID: uuid,
	}
	status, err := h.service.GetBrewStatus(ctx, JobUUIDDTO)
	if err != nil {
		h.sendError(w, err)
		return
	}
	h.sendResponse(w, http.StatusOK, status)
}
