package prompter

import (
	"encoding/json"
)

const (
	parseIngredientDataInstruction = `Make csv with headers: index,basic_ingredient,amount,unit,optional,notes.
Extract the most basic ingredient. Remove cooking tools like spray! Amount must be a number or fraction.
If multiple options given for an ingredient, give the same index, enumerate by letter.
The word or usually indicates multiple options. If no unit or amount, make blank field. Make optional t or f. Use qty for quantity.
Example:
1 pound of beef or pork
1/2 cup of all-purpose flour
Blue food coloring
4 large apples or 2 medium pears,chopped
Can add garlic as garnish
a pinch of salt (optional)

becomes

index,basic_ingredient,amount,unit,optional,notes
1,beef,1,pound,f,
1a,pork,1,pound,f,
2,flour,1/2,cup,f,all-purpose
3,food coloring,,,f,blue
4,apples,4,qty,f,large,chopped
4a,pears,2,qty,f,medium,chopped
5,garlic,,,t,garnish
6,salt,1,pinch,t,
`

	// If multiple options, split them and give the same index, enumerated by lowercase letter, starting with a.

	// vegan, vegetarian, gluten free, dairy free, nut free, shellfish free, egg free, soy free, fish free, pork free, red meat free, alcohol free, kosher, halal
	ingredientAttributePrompt = `Produce a table index,ingredient,breaks, where breaks is a list of the following attributes for the ingredient, and empty if it satisfies everything:
not vegan, not vegetarian, not kosher, not halal, has gluten, has dairy, has nuts, has shellfish, has eggs, has soy, has fish, has pork, has red meat, has alcohol.
Always contain the three columns, even if empty. Evaluate each ingredient separately. Include the header row.
Follow this example:
1,beef
1a,pork
2,flour
3,eggs
4,apple
5,beer

becomes
1,beef,not vegan,not vegetarian,not halal
1a,pork,not vegan,not vegetarian,not kosher,not halal
2,flour,has gluten
3,eggs,not vegan,not vegetarian,has eggs
4,apple,
5,beer,not kosher,not halal,has alcohol`

	parsingModel       = "text-davinci-edit-001"
	catergorizingModel = "text-davinci-003"

	completionURL = "https://api.openai.com/v1/completions"
	editURL       = "https://api.openai.com/v1/edits"
)

type OpenAIRequest interface {
	MakeBody() ([]byte, error)
	URL() string
}

type ParseIngredientsRequest struct {
	ingredients string
}

func NewParseIngredientsRequest(ingredients string) *ParseIngredientsRequest {
	return &ParseIngredientsRequest{
		ingredients: ingredients,
	}
}

func (r *ParseIngredientsRequest) MakeBody() ([]byte, error) {
	reqData := map[string]any{
		"model":       parsingModel,
		"instruction": parseIngredientDataInstruction,
		"input":       r.ingredients,
		"temperature": 0.1,
		"top_p":       0.6,
		"n":           1,
	}

	return json.Marshal(reqData)
}

func (r *ParseIngredientsRequest) URL() string {
	return editURL
}

type IngredientAttributeRequest struct {
	ingredients string
}

func NewIngredientAttributeRequest(ingredients string) *IngredientAttributeRequest {
	return &IngredientAttributeRequest{
		ingredients: ingredients,
	}
}

func (r *IngredientAttributeRequest) MakeBody() ([]byte, error) {
	// reqData := map[string]any{
	// 	"model":       catergorizingModel,
	// 	"prompt":      fmt.Sprintf(ingredientAttributePrompt, r.ingredients),
	// 	"max_tokens":  512,
	// 	"temperature": 0.1,
	// 	"top_p":       0.6,
	// 	"n":           1,
	// 	"stream":      false,
	// 	"logprobs":    nil,
	// }

	reqData := map[string]any{
		"model":       parsingModel,
		"instruction": ingredientAttributePrompt,
		"input":       r.ingredients,
		"temperature": 0.1,
		"top_p":       0.6,
		"n":           1,
	}

	return json.Marshal(reqData)
}

func (r *IngredientAttributeRequest) URL() string {
	// return completionURL
	return editURL
}
