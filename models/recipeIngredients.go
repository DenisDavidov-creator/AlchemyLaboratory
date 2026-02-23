package models

type RecipeIngredients struct {
	RecipeID       int `json:"recipe_id" db:"recipe_id"`
	IngredientID   int `db:"ingredient_id" json:"ingredient_id"`
	QuantityNeeded int `json:"quantity_needed" db:"quantity_needed"`
}
