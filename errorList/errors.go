package errorList

import "errors"

var ErrPostIngredients = errors.New("Error post ingredients")
var ErrAddIngredients = errors.New("Error add ingredients")
var ErrGetIngredients = errors.New("Error get ingredients")
var ErrAddIngredientsNotFound = errors.New("Error add ingredients not found")

var ErrStartTransaction = errors.New("Error start transaction")
var ErrCommitTransaction = errors.New("Error commit transaction")

var ErrCreateRecipe = errors.New("Error create recipe")
var ErrGetRecipes = errors.New("Error get recipes")

var ErrCreateConnectionRecipeIngredient = errors.New("Error create connection recipe ingredient")

var ErrUnmarshal = errors.New("Error unmurshaling")

var ErrWrongJsonFormat = errors.New("Error wrong json format")

var ErrInconsistencyData = errors.New("Error inconsistency data")

var ErrPostJob = errors.New("Error post job")
var ErrSetStatus = errors.New("Error set status")
var ErrGetStatus = errors.New("Error get status")
