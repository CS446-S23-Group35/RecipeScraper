package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/CS446-S23-Group35/RecipeScraper/pkg/recipe"
	css "github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

// The CSS selectors used to parse the Food Network website
const (
	foodnetworkNameSelector        = "span.o-AssetTitle__a-HeadlineText"
	foodNetworkDescriptionSelector = "div.o-AssetDescription__a-Description"
	foodnetworkIngredientsSelector = "p:not(.o-Ingredients__a-Ingredient--SelectAll) span.o-Ingredients__a-Ingredient--CheckboxLabel"
	foodnetworkStepsSelector       = "li.o-Method__m-Step"

	foodnetworkLevelSelector = "div.recipeInfo ul.o-RecipeInfo__m-Level li"
	foodnetworkTimeSelector  = "div.recipeInfo ul.o-RecipeInfo__m-Time li"
	foodnetworkYieldSelector = "div.recipeInfo ul.o-RecipeInfo__m-Yield li"
	foodnetworkTagSelector   = "div.m-TagList a.o-Capsule__a-Tag"
	// Match either normal image or video thumbnail
	foodnetworkImageSelector = "div.m-RecipeMedia__m-MediaBlock img, img.kdp-poster__image"
)

// Some helper selectors, REs and map.
var (
	foodNetworkTimeRE   = regexp.MustCompile(`((\d+) hr)? ?((\d+) min)?$`)
	foodnetworkLevelMap = map[string]recipe.RecipeDifficulty{
		"Easy":         recipe.Easy,
		"Medium":       recipe.Medium,
		"Intermediate": recipe.Medium,
		"Hard":         recipe.Hard,
	}
	foodnetworkYieldRE = regexp.MustCompile(`((\d+)( ?- ?|( to )))?(\d+) servings?`)
	// Match either headline or description spans (click tracking is also a span there).
	foodNetworkMetadataSpanSelector = css.MustCompile("span.o-RecipeInfo__a-Headline, span.o-RecipeInfo__a-Description")
)

// FoodnetworkParser is a parser for the Food Network website. It implements the Parser interface.
// It takes a HTML node and parses it into a RawRecipe struct.
type FoodnetworkParser struct {
	nameSelector        css.Selector
	descriptionSelector css.Selector
	ingredientSelector  css.Selector
	stepsSelector       css.Selector
	levelSelector       css.Selector
	timeSelector        css.Selector
	yieldSelector       css.Selector
	tagSelector         css.Selector
	imageSelector       css.Selector
}

// NewFoodnetworkParser creates a new FoodnetworkParser.
func NewFoodnetworkParser() *FoodnetworkParser {
	return &FoodnetworkParser{
		nameSelector:        css.MustCompile(foodnetworkNameSelector),
		descriptionSelector: css.MustCompile(foodNetworkDescriptionSelector),
		ingredientSelector:  css.MustCompile(foodnetworkIngredientsSelector),
		stepsSelector:       css.MustCompile(foodnetworkStepsSelector),
		levelSelector:       css.MustCompile(foodnetworkLevelSelector),
		timeSelector:        css.MustCompile(foodnetworkTimeSelector),
		yieldSelector:       css.MustCompile(foodnetworkYieldSelector),
		tagSelector:         css.MustCompile(foodnetworkTagSelector),
		imageSelector:       css.MustCompile(foodnetworkImageSelector),
	}
}

// ParseRecipe parses a HTML node into a RawRecipe struct, or returns an error if it fails.
func (p *FoodnetworkParser) ParseRecipe(node *html.Node) (*recipe.RawRecipe, error) {
	name, err := p.parseSingleText(p.nameSelector, node, "name")
	if err != nil {
		return nil, err
	}

	description, err := p.parseSingleText(p.descriptionSelector, node, "description")
	if err != nil {
		// Some recipes don't have a description
		description = ""
	}

	ingredients, err := p.parseListText(p.ingredientSelector, node, "ingredients")
	if err != nil {
		return nil, err
	}

	steps, err := p.parseListText(p.stepsSelector, node, "steps")
	if err != nil {
		return nil, err
	}

	metadata, err := p.parseRecipeMetadata(node)
	if err != nil {
		return nil, err
	}

	return &recipe.RawRecipe{
		Name:                   name,
		Description:            description,
		IngredientDescriptions: ingredients,
		Steps:                  steps,
		Metadata:               metadata,
	}, nil
}

// parseSingleText parses a single text field from a HTML node that gets selected.
func (p *FoodnetworkParser) parseSingleText(selector css.Selector, node *html.Node, name string) (string, error) {
	nodes := selector.MatchAll(node)
	if len(nodes) < 1 {
		return "", ErrParseFailed{Field: name}
	}

	return nodes[0].FirstChild.Data, nil
}

// parseListText parses a list of text fields of the direct children from a HTML node that gets selected.
func (p *FoodnetworkParser) parseListText(selector css.Selector, node *html.Node, name string) ([]string, error) {
	nodes := selector.MatchAll(node)
	if len(nodes) < 1 {
		return nil, ErrParseFailed{Field: name}
	}

	// pre-allocate the slice
	text := make([]string, len(nodes))
	for i, node := range nodes {
		text[i] = node.FirstChild.Data
	}

	return text, nil
}

// parseTimeString returns a x hr y min string as minutes.
func (p *FoodnetworkParser) parseTimeString(text string) (int, error) {
	matches := foodNetworkTimeRE.FindStringSubmatch(text)
	if matches == nil {
		return 0, ErrParseFailed{Field: "time string"}
	}

	// group 2 is hours, group 4 is minutes
	hourStr := matches[2]
	minuteStr := matches[4]

	// if the strings are empty (no match), set them to 0
	if hourStr == "" {
		hourStr = "0"
	}

	if minuteStr == "" {
		minuteStr = "0"
	}

	// know they are ints
	hours, _ := strconv.Atoi(hourStr)
	minutes, _ := strconv.Atoi(minuteStr)

	if hours <= 0 && minutes <= 0 {
		return 0, ErrParseFailed{Field: "time string"}
	}

	// return the minutes
	return hours*60 + minutes, nil
}

// parseServings parses a string like "4 to 6 servings" into a ServingRange struct.
func (p *FoodnetworkParser) parseServings(text string) recipe.ServingRange {
	matches := foodnetworkYieldRE.FindStringSubmatch(text)
	if matches == nil {
		// if there is no match, just return the text as the alternative
		return recipe.ServingRange{Alternative: text}
	}

	// group 2 is min, group 5 is max. If group 2 is empty, set it to group 5
	minStr := matches[2]
	maxStr := matches[5]
	if minStr == "" {
		minStr = maxStr
	}

	// know they are ints
	min, _ := strconv.Atoi(minStr)
	max, _ := strconv.Atoi(maxStr)
	return recipe.ServingRange{Min: min, Max: max}
}

// parseRecipeMetadata parses the metadata of a recipe.
func (p *FoodnetworkParser) parseRecipeMetadata(node *html.Node) (recipe.RecipeMetadata, error) {
	metadata := recipe.RecipeMetadata{EstimatedCalories: -1}

	// parse the difficulty, time, and yield. The place where certain metadata is located is not consistent,
	// e.g. total time is sometimes in Level.
	levelNodes := p.levelSelector.MatchAll(node)
	timeNodes := p.timeSelector.MatchAll(node)
	yieldNodes := p.yieldSelector.MatchAll(node)
	cookingMetadatNodes := append(append(levelNodes, timeNodes...), yieldNodes...)

	// For each node, find the headline and description.
	// Figure out what the headline is, and parse the description accordingly.
	for _, node := range cookingMetadatNodes {
		spans := foodNetworkMetadataSpanSelector.MatchAll(node)
		if len(spans) < 2 {
			continue
		}
		headline, text := spans[0].FirstChild.Data, spans[1].FirstChild.Data

		switch headline {
		case "Level:":
			metadata.Difficulty = foodnetworkLevelMap[text]
		case "Total:":
			time, err := p.parseTimeString(text)
			if err != nil {
				return metadata, err
			}
			metadata.MinutesTotal = time

		case "Prep:":
			time, err := p.parseTimeString(text)
			if err != nil {
				return metadata, err
			}
			metadata.MinutesToPrep = time

		case "Cook:", "Active:":
			time, err := p.parseTimeString(text)
			if err != nil {
				return metadata, err
			}
			metadata.MinutesToCook = time

		case "Yield:":
			metadata.Servings = p.parseServings(text)

		default:
			// Skip unknown headlines
		}
	}

	// Tags are some stuff at the end. Nice to have for extra data about the recipe.
	tagNodes := p.tagSelector.MatchAll(node)
	tags := make([]string, len(tagNodes))
	for i, node := range tagNodes {
		tags[i] = node.FirstChild.Data
	}
	metadata.Tags = tags

	// Parse the image and alt text.
	imageNode := p.imageSelector.MatchFirst(node)
	if imageNode != nil {
		for _, attr := range imageNode.Attr {
			if attr.Key == "src" {
				if !strings.HasSuffix(attr.Val, "1474463768097.jpeg") {
					metadata.ImageURL = "https:" + attr.Val
				}
			}
			if attr.Key == "alt" {
				metadata.ImageAlt = attr.Val
			}
		}
	}

	return metadata, nil
}
