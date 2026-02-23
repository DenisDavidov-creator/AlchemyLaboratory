package boiler

import (
	"alchemicallabaratory/repository"
	"context"
	"time"
)

//go:generate mockery --name=BoilerWorkerInterface
type BoilerWorkerInterface interface {
	Boiled(ctx context.Context, uuid string) error
}

type BoilerWorker struct {
	GrimRepo repository.GrimoireRepoInterface
}

func NewBoilerWorker(repo repository.GrimoireRepoInterface) BoilerWorker {
	return BoilerWorker{
		GrimRepo: repo,
	}
}

func (w BoilerWorker) Boiled(ctx context.Context, uuid string) error {
	timeSeconds, err := w.GrimRepo.GetJobByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(timeSeconds) * time.Second)

	err = w.GrimRepo.SetStatus(ctx, uuid, "completed")
	if err != nil {
		return err
	}
	return nil
}
