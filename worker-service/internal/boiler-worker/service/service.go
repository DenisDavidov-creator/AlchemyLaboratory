package service

import (
	dto "alla/shared/DTO"
	status "alla/shared/status"
	"alla/worker-service/internal/boiler-worker/repository"
	"context"
	"time"
)

//go:generate mockery --name=BoilerWorkerInterface
type ServiceInterface interface {
	Boiled(ctx context.Context, uuid dto.JobUUIDDTO) error
}

type ServiceWorker struct {
	repo repository.RepositoryBrewingInterface
}

func NewBoilerWorker(repo repository.RepositoryBrewingInterface) *ServiceWorker {
	return &ServiceWorker{
		repo: repo,
	}
}

func (w *ServiceWorker) Boiled(ctx context.Context, uuid dto.JobUUIDDTO) error {
	timeSeconds, err := w.repo.StartBrewing(ctx, uuid)
	changeStatus := dto.JobStatusDTO{
		UUID:   uuid.JobUUID,
		Status: status.StatusFailed,
	}

	if err != nil {
		if err := w.repo.SetStatus(ctx, changeStatus); err != nil {
			return err
		}
		return err
	}

	time.Sleep(time.Duration(timeSeconds.BrweingTime) * time.Second)

	changeStatus.Status = status.StatusCompleted
	err = w.repo.SetStatus(ctx, changeStatus)
	if err != nil {
		return err
	}
	return nil
}
