package repository

import (
	dto "alla/shared/DTO"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

//go:generate mockery --name=RepositoryBrewingInterface
type RepositoryBrewingInterface interface {
	GetJobByUUID(context.Context, dto.JobUUIDDTO) (*dto.JobTimeDTO, error)
	SetStatus(context.Context, dto.JobStatusDTO) error
}

type repoBrewing struct {
	DBURL      string
	httpClient *http.Client
}

func NewRepoBrewing(DBURL string) *repoBrewing {
	return &repoBrewing{
		DBURL:      DBURL,
		httpClient: &http.Client{},
	}
}

func (r *repoBrewing) GetJobByUUID(ctx context.Context, req dto.JobUUIDDTO) (*dto.JobTimeDTO, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("GetJobByUUID: %w", err)
	}
	url := fmt.Sprintf("%s/internal/brew", r.DBURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("GetJobByUUID: %w", err)
	}
	httpReq.Header.Set("Content-type", "application/json")

	res, err := r.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("GetJobByUUID: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetJobByUUID: %w", err)
	}

	var time dto.JobTimeDTO

	err = json.NewDecoder(res.Body).Decode(&time)
	if err != nil {
		return nil, fmt.Errorf("GetJobByUUID: %w", err)
	}

	return &time, nil
}
func (r *repoBrewing) SetStatus(ctx context.Context, stutus dto.JobStatusDTO) error {
	body, err := json.Marshal(stutus)
	if err != nil {
		return fmt.Errorf("Repository: %w", err)
	}

	url := fmt.Sprintf("%s/internal/brew/status", r.DBURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("SetStatus: %w", err)
	}
	httpReq.Header.Set("Content-type", "application/json")
	res, err := r.httpClient.Do(httpReq)
	defer res.Body.Close()
	if err != nil {
		return fmt.Errorf("SetStatus: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("SetStatus: %w", err)
	}

	return nil
}
