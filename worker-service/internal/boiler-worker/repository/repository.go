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
	DBURL         string
	httpClient    *http.Client
	brewingClient pb.JobServiceClient
}

func NewRepoBrewing(DBURL string, brewingClient pb.JobServiceClient) *repoBrewing {
	return &repoBrewing{
		DBURL:         DBURL,
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

	// body, err := json.Marshal(req)
	// if err != nil {
	// 	return nil, fmt.Errorf("StartBrewing: %w", err)
	// }
	// url := fmt.Sprintf("%s/internal/brew", r.DBURL)

	// httpReq, err := http.NewRequestWithContext(ctx, "GET", url, bytes.NewBuffer(body))
	// if err != nil {
	// 	return nil, fmt.Errorf("StartBrewing: %w", err)
	// }
	// httpReq.Header.Set("Content-type", "application/json")

	// res, err := r.httpClient.Do(httpReq)
	// if err != nil {
	// 	return nil, fmt.Errorf("StartBrewing: %w", err)
	// }
	// defer res.Body.Close()

	// if res.StatusCode != http.StatusOK {
	// 	return nil, fmt.Errorf("StartBrewing: %w", err)
	// }

	// var time dto.JobTimeDTO

	// err = json.NewDecoder(res.Body).Decode(&time)
	// if err != nil {
	// 	return nil, fmt.Errorf("StartBrewing: %w", err)
	// }

	// return &time, nil
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
	// body, err := json.Marshal(stutus)
	// if err != nil {
	// 	return fmt.Errorf("Repository: %w", err)
	// }

	// url := fmt.Sprintf("%s/internal/brew/status", r.DBURL)

	// httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewBuffer(body))
	// if err != nil {
	// 	return fmt.Errorf("SetStatus: %w", err)
	// }
	// httpReq.Header.Set("Content-type", "application/json")
	// res, err := r.httpClient.Do(httpReq)
	// defer res.Body.Close()
	// if err != nil {
	// 	return fmt.Errorf("SetStatus: %w", err)
	// }
	// if res.StatusCode != http.StatusOK {
	// 	return fmt.Errorf("SetStatus: %w", err)
	// }

	// return nil
}
