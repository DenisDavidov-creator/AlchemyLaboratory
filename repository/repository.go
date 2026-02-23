package repository

import (
	"alchemicallabaratory/errorList"
	"alchemicallabaratory/models"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

//go:generate mockery --name=GrimoireRepoInterface
type GrimoireRepoInterface interface {
	PostIngredients(ctx context.Context, m models.Ingredient) (*models.Ingredient, error)
	AddIngredients(ctx context.Context, id, quantity int) error
	GetIngredients(ctx context.Context) ([]models.Ingredient, error)
	PostRecipe(ctx context.Context, m models.Recipe) (*models.Recipe, error)
	GetRecipes(ctx context.Context) ([]models.Recipe, error)
	PostJob(ctx context.Context, m models.BrewingJobs) (*models.BrewingJobs, error)
	GetJobByUUID(ctx context.Context, uuid string) (int, error)
	GetBrewStatus(ctx context.Context, uuid string) (string, error)
	SetStatus(ctx context.Context, uuid, status string) error
}

type GrimoireRepository struct {
	db *sqlx.DB
}

func NewGrimoireRepository(db *sqlx.DB) *GrimoireRepository {
	return &GrimoireRepository{
		db: db,
	}
}

func (r *GrimoireRepository) PostIngredients(ctx context.Context, m models.Ingredient) (*models.Ingredient, error) {
	var id int

	query := "INSERT INTO ingredients (name, description, quantity) VALUES ($1, $2, $3) RETURNING id"

	err := r.db.GetContext(ctx, &id, query, m.Name, m.Description, m.Quantity)
	if err != nil {
		return nil, errorList.ErrPostIngredients
	}
	m.ID = id
	return &m, nil
}

func (r *GrimoireRepository) AddIngredients(ctx context.Context, id, quantity int) error {

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
		return errorList.ErrAddIngredients
	}

	count, _ := res.RowsAffected()

	if count == 0 {
		return errorList.ErrAddIngredientsNotFound
	}

	return nil
}

func (r *GrimoireRepository) GetIngredients(ctx context.Context) ([]models.Ingredient, error) {
	var allIngredients []models.Ingredient

	query := "SELECT * FROM ingredients"

	err := r.db.SelectContext(ctx, &allIngredients, query)
	if err != nil {
		return nil, errorList.ErrGetIngredients
	}

	return allIngredients, nil
}

func (r *GrimoireRepository) PostRecipe(ctx context.Context, m models.Recipe) (*models.Recipe, error) {
	var recipeID int

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, errorList.ErrStartTransaction
	}
	defer tx.Rollback()

	query := "INSERT INTO recipes (name, description, brewing_time_seconds) VALUES ($1, $2, $3) RETURNING id"

	err = tx.GetContext(ctx, &recipeID, query, m.Name, m.Description, m.BrewingTimeSeconds)
	if err != nil {
		return nil, errorList.ErrCreateRecipe
	}
	m.ID = recipeID

	var valueStrings = []string{}
	var valueArgs = []any{}

	for i, value := range m.Ingredients {
		p1 := i*3 + 1
		p2 := i*3 + 2
		p3 := i*3 + 3
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", p1, p2, p3))
		valueArgs = append(valueArgs, recipeID, value.IngredientID, value.QuantityNeeded)

		m.Ingredients[i].RecipeID = recipeID
	}

	query = fmt.Sprintf("INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity_needed) VALUES %s", strings.Join(valueStrings, ","))

	_, err = tx.ExecContext(ctx, query, valueArgs...)

	if err != nil {
		return nil, errorList.ErrCreateConnectionRecipeIngredient
	}

	if err := tx.Commit(); err != nil {
		return nil, errorList.ErrCommitTransaction
	}

	return &m, nil
}

func (r *GrimoireRepository) GetRecipes(ctx context.Context) ([]models.Recipe, error) {

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
	err := r.db.Select(&rows, query)
	if err != nil {
		return nil, errorList.ErrGetRecipes
	}

	finalRecipes := make([]models.Recipe, len(rows))
	for i := range rows {
		err := json.Unmarshal(rows[i].IngredientsRow, &rows[i].Recipe.Ingredients)
		if err != nil {
			return nil, errorList.ErrUnmarshal
		}
		finalRecipes[i] = rows[i].Recipe
	}

	return finalRecipes, nil
}

func (r *GrimoireRepository) PostJob(ctx context.Context, m models.BrewingJobs) (*models.BrewingJobs, error) {
	var id string
	query := "INSERT INTO brewing_jobs (recipe_id, status, details) VALUES ($1, $2, $3) RETURNING public_id"

	err := r.db.GetContext(ctx, &id, query, m.RecipeID, m.Status, m.Details)
	if err != nil {
		return nil, errorList.ErrPostJob
	}

	m.PublicID = id
	return &m, nil
}

func (r *GrimoireRepository) GetJobByUUID(ctx context.Context, uuid string) (int, error) {

	tx, err := r.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var brewingJob models.BrewingJobs

	query := `
		SELECT 
			id, recipe_id
		FROM 
			brewing_jobs bj 
		WHERE 
			bj.public_id = $1 
			AND bj.status = 'queued' 
		FOR UPDATE
	`

	err = tx.QueryRow(query, uuid).Scan(&brewingJob.ID, &brewingJob.RecipeID)

	if err != nil {
		fmt.Println(1)
		return 0, err
	}

	ingredients := []models.RecipeIngredients{}

	query = `
		SELECT 
			ingredient_id, quantity_needed
		FROM 
			recipe_ingredients
		WHERE 
			recipe_id = $1
	`

	err = tx.Select(&ingredients, query, brewingJob.RecipeID)
	if err != nil {
		fmt.Println(2)
		return 0, err
	}

	for _, value := range ingredients {
		query := `
			UPDATE ingredients 
			SET quantity = quantity - $1
			WHERE id = $2 and quantity >= 0
			RETURNING quantity
		`
		var newQuantity int

		err := tx.QueryRow(query, value.QuantityNeeded, value.IngredientID).Scan(&newQuantity)
		if err != nil {
			fmt.Println(3)
			return 0, err
		}
	}

	query = `
		UPDATE 
			brewing_jobs bj
		SET 
			status = 'processing' 
		WHERE
			bj.id = $1
	`

	tx.Exec(query, brewingJob.ID)

	query = `
		SELECT 
			brewing_time_seconds 
		FROM 
			recipes
		WHERE
			id = $1
	`
	var timeSeconds int

	err = tx.Get(&timeSeconds, query, brewingJob.RecipeID)

	if err != nil {
		fmt.Println(4)
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return timeSeconds, nil
}

func (r *GrimoireRepository) GetBrewStatus(ctx context.Context, uuid string) (string, error) {
	query := `
		SELECT 
			status
		FROM 
			brewing_jobs
		WHERE 
			public_id = $1
	`
	var status string

	err := r.db.Get(&status, query, uuid)
	if err != nil {
		fmt.Println(5)
		return "", errorList.ErrGetStatus
	}
	return status, nil
}

func (r *GrimoireRepository) SetStatus(ctx context.Context, uuid, status string) error {
	query := `
		UPDATE 
			brewing_jobs
		SET
			status = $1,
			completed_at = NOW()
		WHERE 
			public_id = $2
	`

	_, err := r.db.Exec(query, status, uuid)
	if err != nil {
		return errorList.ErrSetStatus
	}
	return nil
}
