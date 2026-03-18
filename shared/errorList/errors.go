package errorList

import "errors"

// PGErrorsCode
const (
	PgUniqueViolation     = "23505"
	PgForeignKeyViolation = "23503"
	PgNotNullViolation    = "23502"
	PgCheckViolation      = "23514"
)

// ingredients
var ErrIngredientNotFound = errors.New("Ingredient not found")
var ErrIngredientAlreadyExist = errors.New("Ingredien aleady exists")

var ErrIngredientNotEnough = errors.New("Ingredient not enough")

// recipes
var ErrRecipeAlreadyExist = errors.New("Recipe already exists")
var ErrRecipeNotFound = errors.New("Recipe not found")

var ErrCreateConnectionRecipeIngredient = errors.New("Error create connection recipe ingredient")

// job
var ErrJobNotFound = errors.New("Job not found")

// requset
var ErrInconsistentData = errors.New("Invalid request data")
var ErrWrongJsonFormat = errors.New("Invalid request format")
