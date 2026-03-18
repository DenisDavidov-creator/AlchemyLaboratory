package handler

import (
	dto "alla/shared/DTO"
	"alla/worker-service/internal/boiler-worker/service"
	"encoding/json"
	"net/http"
)

type HandlerBrewing struct {
	service service.ServiceInterface
}

func NewHandlerBrewing(service service.ServiceInterface) *HandlerBrewing {
	return &HandlerBrewing{
		service: service,
	}
}

func (h *HandlerBrewing) StartBrewing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var uuid dto.JobUUIDDTO
	err := json.NewDecoder(r.Body).Decode(&uuid)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	h.service.Boiled(ctx, uuid)

	w.WriteHeader(http.StatusOK)
}
