package cleaner

import (
	"sort"
	"strings"

	"github.com/CS446-S23-Group35/RecipeScraper/pkg/recipe"
)

type DedupSorter struct {
	nonUnicodeReplace *strings.Replacer
}

func NewDedupSorter() *DedupSorter {
	return &DedupSorter{
		nonUnicodeReplace: strings.NewReplacer(
			" ", " ",
			"’", " ",
		),
	}
}

func (ds *DedupSorter) DedupSort(recipes []recipe.RawRecipe) []recipe.RawRecipe {
	recipeMap := make(map[string]recipe.RawRecipe)
	for _, r := range recipes {
		r.Name = strings.Trim(ds.nonUnicodeReplace.Replace(r.Name), " \n\t")
		r.Description = strings.Trim(ds.nonUnicodeReplace.Replace(r.Description), " \n\t")

		for i, step := range r.Steps {
			nonUnicodeStep := ds.nonUnicodeReplace.Replace(step)
			r.Steps[i] = strings.Trim(nonUnicodeStep, " \n\t")
		}

		for i, ing := range r.IngredientDescriptions {
			nonUnicodeStep := ds.nonUnicodeReplace.Replace(ing)
			r.IngredientDescriptions[i] = strings.Trim(nonUnicodeStep, " \n\t")
		}

		recipeMap[r.Metadata.SourceURL] = r
	}

	i := 0
	for _, r := range recipeMap {
		recipes[i] = r
		i++
	}

	sort.Slice(recipes, func(i, j int) bool {
		return recipes[i].Name < recipes[j].Name
	})

	return recipes
}
