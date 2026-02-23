package models

import (
	dto "alchemicallabaratory/DTO"
	"time"
)

type BrewingJobs struct {
	ID          int
	PublicID    string
	RecipeID    int
	Status      string
	Details     string
	CreatedAt   time.Time
	CompletedAt time.Time
}

func BrewingJobsToModel(recipeID dto.JobDTO) BrewingJobs {
	return BrewingJobs{
		RecipeID:  recipeID.RecipeID,
		Status:    "queued",
		CreatedAt: time.Now(),
	}
}
