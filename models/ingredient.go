package models

import dto "alchemicallabaratory/DTO"

type Ingredient struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Quantity    int    `db:"quantity"`
}

func IngredientsToModel(dto dto.IngredientDTO) Ingredient {
	return Ingredient{
		Name:        dto.Name,
		Description: dto.Description,
		Quantity:    dto.Quantity,
	}
}
