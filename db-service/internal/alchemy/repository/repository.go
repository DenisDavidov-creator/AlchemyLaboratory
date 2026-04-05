package repository

import (
	"alla/db-service/internal/transactor"
	"alla/db-service/models"
	errorList "alla/shared/errorList"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

//go:generate mockery --name=AlchemyRepoInterface
type AlchemyRepoInterface interface {
	CheckIngredientExistsByName(ctx context.Context, name string) (bool, error)
	PostIngredients(ctx context.Context, m models.Ingredient) (*models.Ingredient, error)
	AddIngredients(ctx context.Context, ID, quantity int) error
	GetIngredients(ctx context.Context) ([]models.Ingredient, error)
	GetRecipes(ctx context.Context) ([]models.Recipe, error)
	CreateRecipe(ctx context.Context, m *models.Recipe) error
	CreateRecipeIngredients(ctx context.Context, valueString []string, valueArgs []any) error
	CheckExistRecipeByName(ctx context.Context, name string) (bool, error)
	GetRecipeID(ctx context.Context, uuid string) (int, error)
	GetIngredientsByRecipe(ctx context.Context, recipeID int) ([]models.RecipeIngredients, error)

	CheckingIngridients(ctx context.Context, ings []models.RecipeIngredients) error

	GetBrewingTime(ctx context.Context, recipeId int) (int, error)
}

type AlchemyRepository struct {
	db *sqlx.DB
}

func NewAlchemyRepository(db *sqlx.DB) *AlchemyRepository {
	return &AlchemyRepository{
		db: db,
	}
}

func (r *AlchemyRepository) CheckIngredientExistsByName(ctx context.Context, name string) (bool, error) {

	var exisits bool

	query := `SELECT EXISTS (
		SELECT 1
		FROM ingredients
		WHERE name = $1	
	)`
	err := r.db.GetContext(ctx, &exisits, query, name)

	if err != nil {
		return false, fmt.Errorf("CheckIngredientExistsByName: %w", err)
	}

	return exisits, nil
}

func (r *AlchemyRepository) PostIngredients(ctx context.Context, m models.Ingredient) (*models.Ingredient, error) {
	var id int

	query := "INSERT INTO ingredients (name, description, quantity) VALUES ($1, $2, $3) RETURNING id"

	err := r.db.GetContext(ctx, &id, query, m.Name, m.Description, m.Quantity)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == errorList.PgUniqueViolation {
			return nil, errorList.ErrIngredientAlreadyExist
		}
		return nil, fmt.Errorf("PostIngredients: %w", err)
	}
	m.ID = id
	return &m, nil
}

func (r *AlchemyRepository) AddIngredients(ctx context.Context, id, quantity int) error {

	query := `
		UPDATE
			ingredients
		SET 
			quantity = quantity + $1
		WHERE 
			id = $2
	`

	res, err := r.db.ExecContext(ctx, query, quantity, id)
	if err != nil {
		return fmt.Errorf("AddIngredients: %w", err)
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return errorList.ErrIngredientNotFound
	}

	return nil
}

func (r *AlchemyRepository) GetIngredients(ctx context.Context) ([]models.Ingredient, error) {
	var allIngredients []models.Ingredient

	query := "SELECT * FROM ingredients"

	err := r.db.SelectContext(ctx, &allIngredients, query)
	if err != nil {
		return nil, fmt.Errorf("GetIngredients: %w", err)
	}

	return allIngredients, nil
}

func (r *AlchemyRepository) CheckExistRecipeByName(ctx context.Context, name string) (bool, error) {
	var exist bool
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM 
				recipes 
			WHERE 
				name = $1
		)
	`

	err := r.db.GetContext(ctx, &exist, query, name)

	if err != nil {
		return false, fmt.Errorf("CheckExistRecipeByName: %w", err)
	}
	return exist, nil

}

func (r *AlchemyRepository) CreateRecipe(ctx context.Context, m *models.Recipe) error {
	tx := transactor.GetTx(ctx)
	var recipeID int
	query := "INSERT INTO recipes (name, description, brewing_time_seconds) VALUES ($1, $2, $3) RETURNING id"

	var err error
	if tx != nil {
		err = tx.GetContext(ctx, &recipeID, query, m.Name, m.Description, m.BrewingTimeSeconds)
	} else {
		err = r.db.GetContext(ctx, &recipeID, query, m.Name, m.Description, m.BrewingTimeSeconds)
	}

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == errorList.PgUniqueViolation {
			return errorList.ErrRecipeAlreadyExist
		}
		return fmt.Errorf("CreateRecipe: %w", err)
	}

	m.ID = recipeID
	return nil
}

func (r *AlchemyRepository) CreateRecipeIngredients(ctx context.Context, valueStrings []string, valueArgs []any) error {
	tx := transactor.GetTx(ctx)

	query := fmt.Sprintf("INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity_needed) VALUES %s", strings.Join(valueStrings, ","))

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, valueArgs...)
	} else {
		_, err = r.db.ExecContext(ctx, query, valueArgs...)
	}

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == errorList.PgForeignKeyViolation {
			return errorList.ErrCreateConnectionRecipeIngredient
		}

		return fmt.Errorf("CreateRecipeIngredients: %w", err)
	}
	return nil
}

func (r *AlchemyRepository) GetRecipes(ctx context.Context) ([]models.Recipe, error) {

	type RecipeRow struct {
		models.Recipe
		IngredientsRow []byte `db:"ingredients"`
	}

	rows := []RecipeRow{}

	query := `SELECT r.id, r.name, r.description, r.brewing_time_seconds, 
		COALESCE(
			json_agg(
				json_build_object (
					'recipe_id', ri.recipe_id,
					'ingredient_id', ri.ingredient_id,
					'quantity_needed', ri.quantity_needed			
				)
			) FILTER (WHERE ri.ingredient_id IS NOT NULL), '[]'
		) as ingredients 
		FROM 
			recipes r
		LEFT JOIN recipe_ingredients ri ON r.id = ri.recipe_id
		GROUP BY r.id
	`
	err := r.db.SelectContext(ctx, &rows, query)
	if err != nil {
		return nil, fmt.Errorf("GetRecipes: %w", err)
	}

	finalRecipes := make([]models.Recipe, len(rows))
	for i := range rows {
		err := json.Unmarshal(rows[i].IngredientsRow, &rows[i].Recipe.Ingredients)
		if err != nil {
			return nil, errorList.ErrInconsistentData
		}
		finalRecipes[i] = rows[i].Recipe
	}

	return finalRecipes, nil
}

type GetIDRecipeIDStruct struct {
	RecipeID int `db:"recipe_id"`
}

func (r *AlchemyRepository) GetRecipeID(ctx context.Context, uuid string) (int, error) {

	query := `
		SELECT 
			recipe_id
		FROM 
			brewing_jobs bj 
		WHERE 
			bj.public_id = $1 
			AND bj.status = 'queued' 
		FOR UPDATE
	`
	var data GetIDRecipeIDStruct

	err := r.db.GetContext(ctx, &data, query, uuid)

	if err == sql.ErrNoRows {
		return 0, errorList.ErrRecipeNotFound
	}

	if err != nil {
		return 0, fmt.Errorf("GetRecipeID: %w", err)
	}

	return data.RecipeID, err
}

func (r *AlchemyRepository) GetIngredientsByRecipe(ctx context.Context, recipeID int) ([]models.RecipeIngredients, error) {
	ingredients := []models.RecipeIngredients{}

	tx := transactor.GetTx(ctx)

	query := `
		SELECT 
			ingredient_id, quantity_needed
		FROM 
			recipe_ingredients
		WHERE 
			recipe_id = $1
	`
	var err error
	if tx != nil {
		err = tx.SelectContext(ctx, &ingredients, query, recipeID)
	} else {
		err = r.db.SelectContext(ctx, &ingredients, query, recipeID)
	}

	if err != nil {
		return nil, fmt.Errorf("GetIngredientsByRecipe: %w", err)
	}
	return ingredients, nil
}

func (r *AlchemyRepository) CheckingIngridients(ctx context.Context, ings []models.RecipeIngredients) error {

	tx := transactor.GetTx(ctx)

	for _, value := range ings {
		query := `
			UPDATE ingredients 
			SET quantity = quantity - $1
			WHERE id = $2 and quantity >= $1
			RETURNING quantity
		`
		var newQuantity int
		var err error
		if tx != nil {
			err = tx.QueryRow(query, value.QuantityNeeded, value.IngredientID).Scan(&newQuantity)
		} else {
			err = r.db.QueryRow(query, value.QuantityNeeded, value.IngredientID).Scan(&newQuantity)
		}

		if err != nil {
			if err == sql.ErrNoRows {
				return errorList.ErrIngredientNotEnough
			}
			return fmt.Errorf("CheckingIngridients: %w", err)
		}
	}

	return nil
}

func (r *AlchemyRepository) GetBrewingTime(ctx context.Context, recipeId int) (int, error) {

	query := `
		SELECT 
			brewing_time_seconds 
		FROM 
			recipes
		WHERE
			id = $1
	`
	var timeSeconds int

	err := r.db.GetContext(ctx, &timeSeconds, query, recipeId)

	if err == sql.ErrNoRows {
		return 0, errorList.ErrRecipeNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("GetBrewingTime: %w", err)
	}

	return timeSeconds, nil
}
