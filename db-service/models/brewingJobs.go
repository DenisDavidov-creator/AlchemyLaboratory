package models

import (
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
