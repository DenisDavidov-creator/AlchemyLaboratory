package repository

import (
	dto "alla/shared/DTO"
	"alla/shared/pb"
	"context"
	"fmt"
	"net/http"
)

//go:generate mockery --name=RepositoryBrewingInterface
type RepositoryBrewingInterface interface {
	StartBrewing(context.Context, dto.JobUUIDDTO) (*dto.JobTimeDTO, error)
	SetStatus(context.Context, dto.JobStatusDTO) error
}

type repoBrewing struct {
	httpClient    *http.Client
	brewingClient pb.JobServiceClient
}

func NewRepoBrewing(brewingClient pb.JobServiceClient) *repoBrewing {
	return &repoBrewing{
		httpClient:    &http.Client{},
		brewingClient: brewingClient,
	}
}

func (r *repoBrewing) StartBrewing(ctx context.Context, req dto.JobUUIDDTO) (*dto.JobTimeDTO, error) {

	resp, err := r.brewingClient.StartBrewing(ctx, &pb.JobUUID{JobUUID: req.JobUUID})
	if err != nil {
		return nil, fmt.Errorf("StartBrewing: %w", err)
	}

	return &dto.JobTimeDTO{BrweingTime: int(resp.BrewingTime)}, nil

}
func (r *repoBrewing) SetStatus(ctx context.Context, req dto.JobStatusDTO) error {

	_, err := r.brewingClient.ChangeStatus(ctx, &pb.ChangeJobStatus{
		Uuid:   req.UUID,
		Status: req.Status,
	})
	if err != nil {
		return fmt.Errorf("SetStatus: %w", err)
	}
	return nil

}
