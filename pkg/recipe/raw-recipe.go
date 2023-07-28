package recipe

type RawRecipe struct {
	Name                   string
	Description            string
	IngredientDescriptions []string
	Steps                  []string

	Metadata RecipeMetadata
}

func (r *RawRecipe) ToRecipe() *Recipe {
	return &Recipe{
		Name:        r.Name,
		Description: r.Description,
		Ingredients: nil,
		Steps:       r.Steps,
		Metadata:    r.Metadata,
	}
}
