package service

import (
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	status "alla/shared/status"
	"alla/worker-service/internal/boiler-worker/repository"
	"context"
	"errors"
	"time"
)

//go:generate mockery --name=BoilerWorkerInterface
type ServiceInterface interface {
	Boiled(ctx context.Context, uuid dto.JobUUIDDTO) error
	MarkFailed(ctx context.Context, uuid dto.JobUUIDDTO) error
}

type ServiceWorker struct {
	repo      repository.RepositoryBrewingInterface
	extraWait time.Duration
}

func NewBoilerWorker(repo repository.RepositoryBrewingInterface, extraWait time.Duration) *ServiceWorker {
	return &ServiceWorker{
		repo:      repo,
		extraWait: extraWait,
	}
}

func (w *ServiceWorker) Boiled(ctx context.Context, uuid dto.JobUUIDDTO) error {
	resp, err := w.repo.StartBrewing(ctx, uuid)
	changeStatus := dto.JobStatusDTO{
		UUID:   uuid.JobUUID,
		Status: status.StatusFailed,
	}

	if err != nil {
		if errors.Is(err, errorList.ErrIngredientNotEnough) {
			if err := w.repo.SetStatus(ctx, changeStatus); err != nil {
				return err
			}
		}
		return err
	}

	if resp.Resumed {
		select {
		case <-time.After(time.Duration(resp.BrweingTime)*time.Second + w.extraWait):
			err := w.MarkFailed(ctx, uuid)
			if err != nil && !errors.Is(err, errorList.ErrInvalidStatusTransition) {
				return err
			}
			return errorList.ErrInvalidStatusTransition

		case <-ctx.Done():
			return ctx.Err()
		}
	}

	select {
	case <-time.After(time.Duration(resp.BrweingTime) * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}
	changeStatus.Status = status.StatusCompleted
	err = w.repo.SetStatus(ctx, changeStatus)
	if err != nil {
		return err
	}
	return nil
}

func (w *ServiceWorker) MarkFailed(ctx context.Context, uuid dto.JobUUIDDTO) error {
	changeStatus := dto.JobStatusDTO{
		UUID:   uuid.JobUUID,
		Status: status.StatusFailed,
	}
	err := w.repo.SetStatus(ctx, changeStatus)
	if err != nil {
		return err
	}
	return nil
}
