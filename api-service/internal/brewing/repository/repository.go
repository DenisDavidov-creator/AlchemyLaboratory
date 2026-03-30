package repository

import (
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	"alla/shared/pb"
	"context"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BrewingRepositoryInterface interface {
	PostJob(ctx context.Context, req dto.JobDTO) (*dto.JobUUIDDTO, error)
	Boiled(context.Context, dto.JobUUIDDTO) error
	GetBrewStatus(context.Context, dto.JobUUIDDTO) (*dto.JobStatusresponseDTO, error)
}

type BrewingRepo struct {
	httpClient *http.Client
	jobClient  pb.JobServiceClient
	brewClient pb.BrewServiceClient
}

func NewBrewingRepo(jobClient pb.JobServiceClient, brewClient pb.BrewServiceClient) *BrewingRepo {
	return &BrewingRepo{
		jobClient:  jobClient,
		brewClient: brewClient,

		httpClient: &http.Client{},
	}
}

func (r *BrewingRepo) PostJob(ctx context.Context, req dto.JobDTO) (*dto.JobUUIDDTO, error) {

	resp, err := r.jobClient.PostJob(ctx, &pb.PostJobRequest{
		RecipeId: int32(req.RecipeID),
		Details:  req.Details,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errorList.ErrRecipeNotFound
		}
		return nil, fmt.Errorf("PostJob: %w", err)
	}

	return &dto.JobUUIDDTO{
		JobUUID: resp.JobUUID,
	}, nil

}

func (r *BrewingRepo) Boiled(ctx context.Context, req dto.JobUUIDDTO) error {

	_, err := r.brewClient.Brew(ctx, &pb.JobUUID{JobUUID: req.JobUUID})
	if err != nil {
		return fmt.Errorf("Boild: %w", err)
	}
	return nil

}

func (r *BrewingRepo) GetBrewStatus(ctx context.Context, req dto.JobUUIDDTO) (*dto.JobStatusresponseDTO, error) {

	resp, err := r.jobClient.GetBrewStatus(ctx, &pb.JobUUID{JobUUID: req.JobUUID})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, errorList.ErrJobNotFound
		}
		return nil, fmt.Errorf("GetBrewStatus: %w", err)
	}
	return &dto.JobStatusresponseDTO{Status: resp.Status}, nil

}
