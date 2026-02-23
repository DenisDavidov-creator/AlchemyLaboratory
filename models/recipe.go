package models

import dto "alchemicallabaratory/DTO"

type Recipe struct {
	ID                 int
	Name               string
	Description        string
	BrewingTimeSeconds int `db:"brewing_time_seconds"`
	Ingredients        []RecipeIngredients
}

func RecipeToMdel(rc dto.RecipeDTO) Recipe {

	var ingredients = []RecipeIngredients{}

	for _, value := range rc.Ingredients {
		var ingredient = RecipeIngredients{
			IngredientID:   value.IngredientID,
			QuantityNeeded: value.QuantityNeeded,
		}
		ingredients = append(ingredients, ingredient)
	}

	return Recipe{
		Name:               rc.Name,
		Description:        rc.Description,
		BrewingTimeSeconds: rc.BrewingTimeSeconds,
		Ingredients:        ingredients,
	}
}
