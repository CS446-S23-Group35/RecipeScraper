package recipe

type RawRecipe struct {
	Name                   string
	Description            string
	IngredientDescriptions []string
	Steps                  []string

	Metadata RecipeMetadata
}
