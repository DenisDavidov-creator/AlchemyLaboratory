package dto

type IngredientDTO struct {
	Name        string `json:"name" validate:"required,min=2"`
	Description string `json:"description" validate:"max=500"`
	Quantity    int    `json:"quantity" validate:"gt=0"`
}
type IngredientAddDTO struct {
	Quantity int `json:"quantity" validate:"gt=0"`
}
type RecipeIngredientsDTO struct {
	IngredientID   int `json:"ingredient_id" validate:"required,gt=0"`
	QuantityNeeded int `json:"quantity_needed" validate:"required,gt=0"`
}

type RecipeDTO struct {
	Name               string                 `json:"name" validate:"required,min=2"`
	Description        string                 `json:"description" validate:"max=500"`
	BrewingTimeSeconds int                    `json:"brewing_time_seconds" validate:"gt=0"`
	Ingredients        []RecipeIngredientsDTO `json:"ingredients" validate:"required,min=1,dive"`
}

type JobDTO struct {
	RecipeID int `json:"recipe_id" validate:"gt=0"`
}
