package parser

import (
	"github.com/CS446-S23-Group35/RecipeScraper/pkg/recipe"
	"golang.org/x/net/html"
)

// Parser is an interface for a recipe parser.
// It takes a HTML node and parses it into a RawRecipe struct.
// The ParseRecipe method returns a RawRecipe struct and an error.
type Parser interface {
	ParseRecipe(*html.Node) (*recipe.RawRecipe, error)
}

// ErrParseFailed is an error that is returned when a parser fails to parse a field.
type ErrParseFailed struct {
	Field string
}

func (e ErrParseFailed) Error() string {
	return "failed to parse " + e.Field
}
