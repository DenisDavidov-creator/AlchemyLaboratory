package repository

import (
	dto "alla/shared/DTO"
	"alla/shared/errorList"
	"alla/shared/pb"
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BrewingRepositoryInterface interface {
	PostJob(ctx context.Context, req dto.JobDTO) (*dto.JobUUIDDTO, error)
	Boiled(context.Context, dto.JobUUIDDTO) error
	GetBrewStatus(context.Context, dto.JobUUIDDTO) (*dto.JobStatusresponseDTO, error)
}

type BrewingRepo struct {
	jobClient   pb.JobServiceClient
	kafkaClient *kgo.Client
}

func NewBrewingRepo(jobClient pb.JobServiceClient, kafkaClient *kgo.Client) *BrewingRepo {
	return &BrewingRepo{
		jobClient:   jobClient,
		kafkaClient: kafkaClient,
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
