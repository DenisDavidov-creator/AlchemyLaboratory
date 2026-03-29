package handler

import (
	"alla/db-service/internal/brewing/service"
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type BrewingHandler struct {
	service service.BrewingServiceInterface
}

func NewBrewingHandler(service service.BrewingServiceInterface) *BrewingHandler {
	return &BrewingHandler{
		service: service,
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
	}{
		Message: fmt.Sprintf("DB-serivce: %v", err),
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

func (h *BrewingHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req dto.JobDTO

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		fmt.Println("Wrong JSON format")
		h.sendError(w, fmt.Errorf("Handler : %w", err))

	}

	res, err := h.service.CreateJob(ctx, req)
	if err != nil {
		h.sendError(w, fmt.Errorf("Handler : %w", err))
	}
	h.sendResponse(w, http.StatusCreated, res)

}

func (h *BrewingHandler) StartBrewing(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	var req dto.JobUUIDDTO

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.sendError(w, fmt.Errorf("Handler: %w", err))
	}

	res, err := h.service.StartBrewing(ctx, req)

	if err != nil {
		h.sendError(w, fmt.Errorf("Handler : %w", err))
	}
	h.sendResponse(w, http.StatusOK, res)
}

func (h *BrewingHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req dto.JobStatusDTO
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.sendError(w, fmt.Errorf("Handler: %w", err))
	}

	err = h.service.SetStatus(ctx, req)

	if err != nil {
		h.sendError(w, fmt.Errorf("Handler : %w", err))
	}
	h.sendResponse(w, http.StatusOK, nil)
}

func (h *BrewingHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req dto.JobUUIDDTO

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.sendError(w, fmt.Errorf("Handler: %w", err))
	}

	res, err := h.service.GetBrewStatus(ctx, req)
	if err != nil {
		h.sendError(w, fmt.Errorf("Handler : %w", err))
	}

	h.sendResponse(w, http.StatusOK, res)
}
