package parser

import (
	"strings"

	"github.com/CS446-S23-Group35/RecipeScraper/pkg/recipe"
	css "github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

const (
	wikibooksNameSelector        = "h1#firstHeading"
	wikibooksIngredientsSelector = "ul li:has(a[title^='Cookbook:'])"
	wikibooksStepsSelector       = "ol.li.a[title]"
)

type WikibooksParser struct {
	nameSelector       css.Selector
	ingredientSelector css.Selector
	stepsSelector      css.Selector
}

func NewWikibooksParser() *WikibooksParser {
	// Compile the required css selectors, panic if any fail
	nameSelector := css.MustCompile(wikibooksNameSelector)
	ingredientSelector := css.MustCompile(wikibooksIngredientsSelector)
	stepsSelector := css.MustCompile(wikibooksStepsSelector)

	return &WikibooksParser{
		nameSelector:       nameSelector,
		ingredientSelector: ingredientSelector,
		stepsSelector:      stepsSelector,
	}
}

func (p *WikibooksParser) ParseRecipe(node *html.Node) (*recipe.RawRecipe, error) {
	name, err := p.parseName(node)
	if err != nil {
		return nil, err
	}

	// ings := p.ingredientSelector.MatchAll(node)
	// for _, ing := range ings {
	// 	html.Render(os.Stdout, ing)
	// }

	return &recipe.RawRecipe{
		Name: name,
	}, nil
}

func (p *WikibooksParser) parseName(node *html.Node) (string, error) {
	titleNode := p.nameSelector.MatchAll(node)
	if len(titleNode) != 1 {
		return "", ErrParseFailed{Field: "name"}
	}

	title := titleNode[0].FirstChild.Data

	return strings.TrimPrefix(title, "Cookbook:"), nil
}
