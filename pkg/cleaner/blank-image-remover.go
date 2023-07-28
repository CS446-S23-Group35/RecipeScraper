package cleaner

import "github.com/CS446-S23-Group35/RecipeScraper/pkg/recipe"

type BlankRemover struct {
}

func NewBlankRemover() *BlankRemover {
	return &BlankRemover{}
}

func (bir *BlankRemover) RemoveBlankImages(recipes []recipe.RawRecipe) []recipe.RawRecipe {
	return bir.removeBlank(recipes, func(r recipe.RawRecipe) bool {
		return r.Metadata.ImageURL == ""
	})
}

func (bir *BlankRemover) RemoveBlankDescription(recipes []recipe.RawRecipe) []recipe.RawRecipe {
	return bir.removeBlank(recipes, func(r recipe.RawRecipe) bool {
		return r.Description == ""
	})
}

func (bir *BlankRemover) removeBlank(recipes []recipe.RawRecipe, isBlank func(recipe.RawRecipe) bool) []recipe.RawRecipe {
	newRecipes := make([]recipe.RawRecipe, 0, len(recipes))
	for _, r := range recipes {
		if !isBlank(r) {
			newRecipes = append(newRecipes, r)
		}
	}
	return newRecipes
}
