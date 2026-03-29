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
	dbUrl      string
	workerUrl  string
	httpClient *http.Client
	jobClient  pb.JobServiceClient
	brewClient pb.BrewServiceClient
}

func NewBrewingRepo(dbUrl string, workerUrl string, jobClient pb.JobServiceClient, brewClient pb.BrewServiceClient) *BrewingRepo {
	return &BrewingRepo{
		dbUrl:      dbUrl,
		workerUrl:  workerUrl,
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

	// body, err := json.Marshal(req)
	// if err != nil {
	// 	return nil, fmt.Errorf("Repository: %w", err)
	// }

	// url := fmt.Sprintf("%s/internal/brew", r.dbUrl)
	// httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))

	// httpReq.Header.Set("Content-type", "application/json")

	// resp, err := r.httpClient.Do(httpReq)
	// if err != nil {
	// 	return nil, err
	// }
	// defer resp.Body.Close()

	// switch resp.StatusCode {
	// case http.StatusCreated:
	// 	var result dto.JobUUIDDTO
	// 	err = json.NewDecoder(resp.Body).Decode(&result)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return &result, nil
	// case http.StatusNotFound:
	// 	return nil, errorList.ErrRecipeNotFound
	// case http.StatusUnprocessableEntity:
	// 	return nil, errorList.ErrJobNotFound
	// default:
	// 	return nil, fmt.Errorf("PostJob: unexpected error: %d", resp.StatusCode)
	// }

}

func (r *BrewingRepo) Boiled(ctx context.Context, req dto.JobUUIDDTO) error {

	_, err := r.brewClient.Brew(ctx, &pb.JobUUID{JobUUID: req.JobUUID})
	if err != nil {
		return fmt.Errorf("Boild: %w", err)
	}
	return nil
	// body, err := json.Marshal(req)
	// if err != nil {
	// 	return fmt.Errorf("BrewingRepository: %w", err)
	// }

	// url := fmt.Sprintf("%s/internal/brew", r.workerUrl)
	// httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))

	// httpReq.Header.Set("Content-type", "application/json")

	// res, err := r.httpClient.Do(httpReq)
	// if err != nil {
	// 	return fmt.Errorf("BrewingRepository: %w", err)
	// }

	// if res.StatusCode != http.StatusOK {
	// 	return fmt.Errorf("BrewingRepository: %w", err)
	// }
	// return nil
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

	// body, err := json.Marshal(req)
	// if err != nil {
	// 	return nil, fmt.Errorf("BrewingRepository: %w", err)
	// }

	// url := fmt.Sprintf("%s/internal/brew/status", r.dbUrl)
	// httpReq, err := http.NewRequestWithContext(ctx, "GET", url, bytes.NewBuffer(body))

	// httpReq.Header.Set("Content-type", "application/json")

	// res, err := r.httpClient.Do(httpReq)
	// if err != nil {
	// 	return nil, fmt.Errorf("BrewingRepository: %w", err)
	// }

	// switch res.StatusCode {
	// case http.StatusOK:
	// 	var JobUUIDDTO dto.JobStatusresponseDTO
	// 	err = json.NewDecoder(res.Body).Decode(&JobUUIDDTO)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("BrewingRepository: %w", err)
	// 	}

	// 	return &JobUUIDDTO, nil
	// case http.StatusNotFound:
	// 	return nil, errorList.ErrRecipeNotFound
	// default:
	// 	return nil, fmt.Errorf("GetBrewStatus: unexpected status code: %d", res.StatusCode)
	// }
}
