package service

import (
	"alla/api-service/internal/brewing/repository"
	dto "alla/shared/DTO"
	"context"
	"log"
)

//go:generate mockery --name=BrewingServiceInterface
type BrewingServiceInterface interface {
	PostJob(ctx context.Context, jobDTO dto.JobDTO) (*dto.JobUUIDDTO, error)
	Boiled(ctx context.Context, JobUUIDDTO *dto.JobUUIDDTO) error
	GetBrewStatus(ctx context.Context, JobUUIDDTO dto.JobUUIDDTO) (*dto.JobStatusresponseDTO, error)
}

type BrewingService struct {
	repo repository.BrewingRepositoryInterface
}

func NewBrewingService(repo repository.BrewingRepositoryInterface) *BrewingService {
	return &BrewingService{
		repo: repo,
	}
}

func (s *BrewingService) PostJob(ctx context.Context, jobDTO dto.JobDTO) (*dto.JobUUIDDTO, error) {
	uuid, err := s.repo.PostJob(ctx, jobDTO)
	if err != nil {
		return nil, err
	}

	go func() {
		err := s.Boiled(context.Background(), uuid)
		if err != nil {
			log.Printf("Boiled error: %v", err)
		}
	}()

	return uuid, nil
}

func (s *BrewingService) Boiled(ctx context.Context, JobUUIDDTO *dto.JobUUIDDTO) error {

	err := s.repo.Boiled(ctx, *JobUUIDDTO)

	if err != nil {
		return err
	}
	return nil
}

func (s *BrewingService) GetBrewStatus(ctx context.Context, JobUUIDDTO dto.JobUUIDDTO) (*dto.JobStatusresponseDTO, error) {
	status, err := s.repo.GetBrewStatus(ctx, JobUUIDDTO)
	if err != nil {
		return nil, err
	}
	return status, nil

}
