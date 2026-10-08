package service

import (
	alchemyRepo "alla/db-service/internal/alchemy/repository"
	brewingRepo "alla/db-service/internal/brewing/repository"
	"alla/db-service/internal/transactor"
	"alla/db-service/models"
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	status "alla/shared/status"
	"context"
	"fmt"
)

type BrewingServiceInterface interface {
	CreateJob(ctx context.Context, m dto.JobDTO) (*dto.JobUUIDDTO, error)
	StartBrewing(ctx context.Context, uuid dto.JobUUIDDTO) (*dto.JobTimeDTO, error)
	GetBrewStatus(ctx context.Context, uuid dto.JobUUIDDTO) (*dto.JobStatusresponseDTO, error)
	SetStatus(ctx context.Context, req dto.JobStatusDTO) error
}

type BrewingService struct {
	BrewingRepo brewingRepo.BrewingRepoInterface
	AlchemyRepo alchemyRepo.AlchemyRepoInterface
	Trans       transactor.TransactorInterface
}

func NewBrewingService(BrewingRepo brewingRepo.BrewingRepoInterface, AlchemyRepo alchemyRepo.AlchemyRepoInterface, Trans transactor.TransactorInterface) *BrewingService {
	return &BrewingService{
		BrewingRepo: BrewingRepo,
		AlchemyRepo: AlchemyRepo,
		Trans:       Trans,
	}
}

func (s *BrewingService) CreateJob(ctx context.Context, m dto.JobDTO) (*dto.JobUUIDDTO, error) {

	jobModel := models.BrewingJobs{
		RecipeID: m.RecipeID,
		Status:   status.StatusQueued,
		Details:  m.Details,
	}
	err := s.BrewingRepo.CreateJob(ctx, &jobModel)
	if err != nil {
		return nil, fmt.Errorf("CreateJob: %w", err)
	}

	UUID := dto.JobUUIDDTO{
		JobUUID: jobModel.PublicID,
	}
	return &UUID, nil
}

func (s *BrewingService) StartBrewing(ctx context.Context, uuid dto.JobUUIDDTO) (*dto.JobTimeDTO, error) {

	var brewingTime int
	recipeId, err := s.AlchemyRepo.GetRecipeID(ctx, uuid.JobUUID)
	if err != nil {
		return nil, fmt.Errorf("StartBrewing: %w", err)
	}

	ingredients, err := s.AlchemyRepo.GetIngredientsByRecipe(ctx, recipeId)
	if err != nil {
		return nil, fmt.Errorf("StartBrewing: %w", err)
	}

	brewingTime, err = s.AlchemyRepo.GetBrewingTime(ctx, recipeId)
	if err != nil {
		return nil, fmt.Errorf("StartBrewing: %w", err)
	}

	allowedTransitions, err := CheckStatus(status.StatusBrewing)
	if err != nil {
		return nil, fmt.Errorf("service: %w", err)
	}

	err = s.Trans.WithinTransaction(ctx, func(ctx context.Context) error {
		err = s.BrewingRepo.SetStatus(ctx, uuid.JobUUID, status.StatusBrewing, allowedTransitions)
		if err != nil {
			return err
		}

		err = s.AlchemyRepo.CheckingIngridients(ctx, ingredients)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("StartBrewing: %w", err)
	}

	return &dto.JobTimeDTO{BrweingTime: brewingTime}, nil
}

func (s *BrewingService) GetBrewStatus(ctx context.Context, uuid dto.JobUUIDDTO) (*dto.JobStatusresponseDTO, error) {

	status, err := s.BrewingRepo.GetBrewStatus(ctx, uuid.JobUUID)
	if err != nil {
		return nil, err
	}

	res := dto.JobStatusresponseDTO{
		Status: status,
	}

	return &res, nil
}

func (s *BrewingService) SetStatus(ctx context.Context, req dto.JobStatusDTO) error {

	allowedTransitions, err := CheckStatus(req.Status)
	if err != nil {
		return fmt.Errorf("service: %w", err)
	}
	err = s.BrewingRepo.SetStatus(ctx, req.UUID, req.Status, allowedTransitions)
	if err != nil {
		return fmt.Errorf("service: %w", err)
	}
	return nil
}

func CheckStatus(wantedStatus string) ([]string, error) {
	var allowedTransitions = map[string][]string{
		status.StatusBrewing:   {status.StatusQueued},
		status.StatusCompleted: {status.StatusBrewing},
		status.StatusFailed:    {status.StatusBrewing, status.StatusQueued},
	}
	from, ok := allowedTransitions[wantedStatus]
	if !ok {
		return nil, errorList.ErrUnexpectedStatus
	}
	return from, nil

}
