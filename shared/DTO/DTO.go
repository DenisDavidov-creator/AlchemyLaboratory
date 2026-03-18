package dto

type IngredientDTO struct {
	Name        string `json:"name" validate:"required,min=2"`
	Description string `json:"description" validate:"max=500"`
	Quantity    int    `json:"quantity" validate:"gt=0"`
}
type IngredientAddDTO struct {
	Quantity int `json:"quantity" validate:"gt=0"`
}

type IngredientResponseDTO struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
}

type IngredientIDResponseDTO struct {
	ID int `json:"id"`
}

type UpdateIngredientQuantityDTO struct {
	ID       int `json:"-"`
	Quantity int `json:"quantity"`
}

type RecipeIngredientsDTO struct {
	IngredientID   int `json:"ingredient_id" validate:"required,gt=0"`
	QuantityNeeded int `json:"quantity_needed" validate:"required,gt=0"`
}

type RecipeIDResponseDTO struct {
	ID int `json:"id"`
}

type RecipeResponseDTO struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	BrewingTimeSeconds int    `json:"brewing_time_seconds"`
	Ingredients        []RecipeIngredientsDTO
}

type RecipeDTO struct {
	Name               string                 `json:"name" validate:"required,min=2"`
	Description        string                 `json:"description" validate:"max=500"`
	BrewingTimeSeconds int                    `json:"brewing_time_seconds" validate:"gt=0"`
	Ingredients        []RecipeIngredientsDTO `json:"ingredients" validate:"required,min=1,dive"`
}

type JobDTO struct {
	RecipeID int    `json:"recipe_id" validate:"gt=0"`
	Details  string `json:"details"`
}

type JobUUIDDTO struct {
	JobUUID string `json:"job_uuid"`
}

type JobStatusresponseDTO struct {
	Status string `json:"status"`
}

type JobTimeDTO struct {
	BrweingTime int `json:"brewing_time"`
}

type JobStatusDTO struct {
	UUID   string `json:"uuid"`
	Status string `json:"status"`
}
