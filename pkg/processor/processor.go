package processor

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/CS446-S23-Group35/RecipeScraper/pkg/prompter"
	"github.com/CS446-S23-Group35/RecipeScraper/pkg/recipe"
	"gopkg.in/yaml.v3"
)

type RecipeProcessor struct {
	prompter    prompter.OpenAIPrompter
	logFile     *os.File
	successFile *os.File
	outputFile  *os.File

	logMutex     sync.Mutex
	successMutex sync.Mutex
	outputMutex  sync.Mutex
}

func (p *RecipeProcessor) writeMsg(msg string) {
	p.logMutex.Lock()
	defer p.logMutex.Unlock()
	fmt.Fprintf(p.logFile, "INFO::%s\n", msg)
}

func (p *RecipeProcessor) writeErr(recipe *recipe.RawRecipe, err error, workerNum int) {
	p.logMutex.Lock()
	defer p.logMutex.Unlock()
	fmt.Fprintf(p.logFile, "ERROR::%d: processing recipe %s: %s\n", workerNum, recipe.Name, err.Error())
}

func (p *RecipeProcessor) writeRecErr(recipe *recipe.Recipe, err error, workerNum int) {
	p.logMutex.Lock()
	defer p.logMutex.Unlock()
	fmt.Fprintf(p.logFile, "ERROR::%d: processing recipe %s: %s\n", workerNum, recipe.Name, err.Error())
}

func (p *RecipeProcessor) writeSuccess(recipe *recipe.RawRecipe, workerNum int) {
	p.successMutex.Lock()
	defer p.successMutex.Unlock()
	fmt.Fprintf(p.successFile, "%d:%s::%s\n", workerNum, recipe.Name, recipe.Metadata.SourceURL)
}

func (p *RecipeProcessor) writeRecSuccess(recipe *recipe.Recipe, workerNum int) {
	p.successMutex.Lock()
	defer p.successMutex.Unlock()
	fmt.Fprintf(p.successFile, "%d:%s::%s\n", workerNum, recipe.Name, recipe.Metadata.SourceURL)
}

func (p *RecipeProcessor) writeOutput(recipeOut *recipe.Recipe) {
	out := []*recipe.Recipe{recipeOut}
	p.outputMutex.Lock()
	defer p.outputMutex.Unlock()
	err := yaml.NewEncoder(p.outputFile).Encode(out)
	if err != nil {
		raw := &recipe.RawRecipe{Metadata: recipeOut.Metadata}
		p.writeErr(raw, fmt.Errorf("error encoding: %w", err), -1)
	}
}

func NewRecipeProcessor() *RecipeProcessor {
	logFile, err := os.Create("logs/processor.log")
	if err != nil {
		panic(err)
	}
	successFile, err := os.OpenFile("logs/processor_success.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	timeStamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("recipes/ing_proc_recipes_%s.yaml", timeStamp)
	outputFile, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	return &RecipeProcessor{
		prompter:    *prompter.NewOpenAIPrompter(),
		logFile:     logFile,
		successFile: successFile,
		outputFile:  outputFile,
	}
}

func (p *RecipeProcessor) Close() {
	p.logFile.Close()
	p.successFile.Close()
}

func (p *RecipeProcessor) ProcessRawRecipes(recipes []*recipe.RawRecipe, n int, workers int) error {
	recipeChan := make(chan *recipe.RawRecipe, n)

	rand.Shuffle(len(recipes), func(i, j int) { recipes[i], recipes[j] = recipes[j], recipes[i] })
	recipes = recipes[:n]

	wg := sync.WaitGroup{}
	wg.Add(workers)

	// producer
	go func() {
		for _, recipeIn := range recipes {
			recipeChan <- recipeIn
		}
	}()

	for i := 0; i < workers; i++ {
		go func(i int) {
			for recipeIn := range recipeChan {
				processedRecipes, err := p.ProcessRecipe(recipeIn, i)
				if err != nil {
					p.writeErr(recipeIn, err, i)
					continue
				}
				for _, recipeOut := range processedRecipes {
					if recipeOut == nil {
						continue
					}
					p.writeOutput(recipeOut)
				}
				p.writeSuccess(recipeIn, i)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return nil
}

func (p *RecipeProcessor) ProcessRecipeAttributes(recipes map[string][]*recipe.Recipe, n int, workers int) error {
	recipeChan := make(chan []*recipe.Recipe, n)

	wg := sync.WaitGroup{}
	wg.Add(workers)

	// producer
	go func() {
		for _, recipeIn := range recipes {
			recipeChan <- recipeIn
		}
	}()

	for i := 0; i < workers; i++ {
		go func(i int) {
			for recipeIn := range recipeChan {
				processedRecipes, err := p.ProcessAttributes(recipeIn, i)
				if err != nil {
					p.writeRecErr(recipeIn[0], err, i)
					continue
				}
				for _, recipeOut := range processedRecipes {
					if recipeOut == nil {
						continue
					}
					p.writeOutput(recipeOut)
				}
				p.writeRecSuccess(recipeIn[0], i)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return nil
}

func (p *RecipeProcessor) ProcessAttributes(groupedByUrl []*recipe.Recipe, workerNum int) ([]*recipe.Recipe, error) {
	fmt.Println("Processing attributes for " + groupedByUrl[0].Name)
	combinedIng := make(map[string]bool)
	for _, recipe := range groupedByUrl {
		for _, ingredient := range recipe.Ingredients {
			combinedIng[ingredient.Name] = true
		}
	}

	ingredients := make([]string, 0, len(combinedIng))
	for ingredient := range combinedIng {
		ingredients = append(ingredients, ingredient)
	}

	ingredientsStr := ""
	ingMap := make(map[string]int)
	for i, ingredient := range ingredients {
		ingredientsStr += fmt.Sprintf("%d, %s\n", i, ingredient)
		ingMap[ingredient] = i
	}

	p.writeMsg(fmt.Sprintf("%d: Input:\n%s", workerNum, ingredientsStr))

	request := prompter.NewIngredientAttributeRequest(ingredientsStr)
	resp, err := p.prompter.MakeRequest(request)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from openai for attributes")
	}

	p.writeMsg(fmt.Sprintf("%d: Parsed attributes:\n%s", workerNum, resp.Choices[0].Text))

	dietary, err := p.parseDietary(resp.Choices[0].Text)
	if err != nil {
		return nil, err
	}

	out := make([]*recipe.Recipe, len(groupedByUrl))
	copy(out, groupedByUrl)

	for _, recipe := range out {
		for _, ingredient := range recipe.Ingredients {
			for _, label := range dietary[ingredient.Name] {
				switch strings.ToLower(label) {
				case "not vegan":
					recipe.Metadata.Dietary.IsVegan = false
				case "not vegetarian":
					recipe.Metadata.Dietary.IsVegetarian = false
				case "has gluten":
					recipe.Metadata.Dietary.IsGlutenFree = false
				case "has dairy":
					recipe.Metadata.Dietary.IsDairyFree = false
				case "has nuts":
					recipe.Metadata.Dietary.IsNutFree = false
				case "has shellfish":
					recipe.Metadata.Dietary.IsShellfishFree = false
				case "has eggs":
					recipe.Metadata.Dietary.IsEggFree = false
				case "has soy":
					recipe.Metadata.Dietary.IsSoyFree = false
				case "has fish":
					recipe.Metadata.Dietary.IsFishFree = false
				case "has pork":
					recipe.Metadata.Dietary.IsPorkFree = false
				case "has red meat":
					recipe.Metadata.Dietary.IsRedMeatFree = false
				case "has alcohol":
					recipe.Metadata.Dietary.IsAlcoholFree = false
				case "not kosher":
					recipe.Metadata.Dietary.IsKosher = false
				case "not halal":
					recipe.Metadata.Dietary.IsHalal = false
				default:
				}
			}
		}
	}
	return out, nil
}

func (p *RecipeProcessor) ProcessRecipe(recipeIn *recipe.RawRecipe, workerNum int) ([]*recipe.Recipe, error) {
	fmt.Printf("Processing: %s, %s\n", recipeIn.Name, recipeIn.Metadata.SourceURL)
	ingredients := p.reorderIngredients(recipeIn.IngredientDescriptions)
	ingredientsStr := ""
	for _, ingredient := range ingredients {
		ingredientsStr += ingredient + "\n"
	}

	request := prompter.NewParseIngredientsRequest(ingredientsStr)
	resp, err := p.prompter.MakeRequest(request)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from openai for ingredients")
	}

	p.writeMsg(fmt.Sprintf("%d: Parsed ingredients %s:\n%s", workerNum, recipeIn.Name, resp.Choices[0].Text))

	parsedIngredients, err := p.parseIngredients(resp.Choices[0].Text)
	if err != nil {
		return nil, err
	}

	dietaryReqInput := "index,ingredient\n"
	for _, parsedIngredient := range parsedIngredients {
		dietaryReqInput += fmt.Sprintf("%s,%s\n", parsedIngredient.Index, parsedIngredient.Ingredient)
	}

	// dietaryReq := prompter.NewIngredientAttributeRequest(dietaryReqInput)
	// resp2, err := p.prompter.MakeRequest(dietaryReq)
	// if err != nil {
	// 	return nil, err
	// }

	// if len(resp2.Choices) == 0 {
	// 	return nil, fmt.Errorf("no response from openai for dietary")
	// }

	// p.writeMsg(fmt.Sprintf("Dietary:\n%s", resp2.Choices[0].Text))
	// dietary, err := p.parseDietary(resp2.Choices[0].Text)
	// if err != nil {
	// 	return nil, err
	// }

	dietary := make(map[string][]string)

	variants := p.createIndexMap(parsedIngredients)

	recipeOut := make([]*recipe.Recipe, len(variants))

	for _, variant := range variants {
		variantDietary := recipe.RecipeDietaryInformation{
			IsVegan:         true,
			IsVegetarian:    true,
			IsGlutenFree:    true,
			IsDairyFree:     true,
			IsNutFree:       true,
			IsShellfishFree: true,
			IsEggFree:       true,
			IsSoyFree:       true,
			IsFishFree:      true,
			IsPorkFree:      true,
			IsRedMeatFree:   true,
			IsAlcoholFree:   true,
			IsKosher:        true,
			IsHalal:         true,
		}

		for _, index := range variant {
			for _, label := range dietary[parsedIngredients[index].Index] {
				switch strings.ToLower(label) {
				case "not vegan":
					variantDietary.IsVegan = false
				case "not vegetarian":
					variantDietary.IsVegetarian = false
				case "has gluten":
					variantDietary.IsGlutenFree = false
				case "has dairy":
					variantDietary.IsDairyFree = false
				case "has nuts":
					variantDietary.IsNutFree = false
				case "has shellfish":
					variantDietary.IsShellfishFree = false
				case "has eggs":
					variantDietary.IsEggFree = false
				case "has soy":
					variantDietary.IsSoyFree = false
				case "has fish":
					variantDietary.IsFishFree = false
				case "has pork":
					variantDietary.IsPorkFree = false
				case "has red meat":
					variantDietary.IsRedMeatFree = false
				case "has alcohol":
					variantDietary.IsAlcoholFree = false
				case "not kosher":
					variantDietary.IsKosher = false
				case "not halal":
					variantDietary.IsHalal = false
				default:
				}
			}
		}

		ingredients := make(recipe.IngredientList, len(variant))
		for i, index := range variant {
			unitStr := parsedIngredients[index].Unit
			unit := recipe.UnitFromStr(unitStr)
			frac, isFrac := recipe.ParseFraction(parsedIngredients[index].Amount)
			if frac == -1 {
				frac = 1
			}
			if isFrac && unit == recipe.UnitQuanity {
				unit = recipe.UnitFraction
			}

			ingredients[i] = recipe.IngredientItem{
				Name: parsedIngredients[index].Ingredient,
				Amount: recipe.Amount{
					Value:    frac,
					Type:     unit,
					TypeName: unitStr,
				},
				Optional: strings.Contains(strings.ToLower(parsedIngredients[index].Optional), "t"),
				Notes:    parsedIngredients[index].Notes,
			}
		}

		recipeResult := recipeIn.ToRecipe()
		recipeResult.Ingredients = ingredients
		recipeResult.Metadata.Dietary = variantDietary

		recipeOut = append(recipeOut, recipeResult)
	}

	return recipeOut, nil
}

// For each index, creats a list of labels for each dietary restriction that
// the ingredient breaks
func (p *RecipeProcessor) parseIngredients(s string) ([]Ingredient, error) {
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	ind := 1
	if lines[0] != "index,basic_ingredient,amount,unit,optional,notes" {
		if lines[0][0] == '1' {
			ind = 0
		} else {
			return nil, fmt.Errorf("invalid ingredient response: bad header")
		}
	}

	ingredients := make([]Ingredient, 0)
	for _, line := range lines[ind:] {
		parts := strings.Split(line, ",")
		if len(parts) < 6 {
			return nil, fmt.Errorf("invalid ingredient response: bad line %s", line)
		}

		ingredients = append(ingredients, Ingredient{
			Index:      parts[0],
			Ingredient: parts[1],
			Amount:     parts[2],
			Unit:       parts[3],
			Optional:   parts[4],
			Notes:      strings.Join(parts[5:], ","),
		})
	}

	return ingredients, nil
}

// For each index, creats a list of labels for each dietary restriction that
// the ingredient breaks
func (p *RecipeProcessor) parseDietary(s string) (map[string][]string, error) {
	s = strings.TrimSpace(s)
	lines := strings.Split(s, "\n")
	ind := 1
	if lines[0] != "index,ingredient,breaks" {
		if lines[0][0] == '0' {
			ind = 0
		} else {
			return nil, fmt.Errorf("invalid dietary response: bad header")
		}
	}

	dietary := make(map[string][]string)
	for _, line := range lines[ind:] {
		parts := strings.Split(line, ",")
		if len(parts) == 0 {
			return nil, fmt.Errorf("invalid dietary response: bad line %s", line)
		}
		for i, label := range parts[2:] {
			parts[i+2] = strings.TrimSpace(label)
		}
		dietary[parts[1]] = parts[2:]
	}

	return dietary, nil
}

func (p *RecipeProcessor) reorderIngredients(ingredients []string) []string {
	i := 0
	j := len(ingredients) - 1
	for i < j {
		if strings.Contains(ingredients[i], " or ") {
			i += 1
		} else if strings.Contains(ingredients[j], " or ") {
			ingredients[i], ingredients[j] = ingredients[j], ingredients[i]
			i += 1
			j -= 1
		} else {
			j -= 1
		}
	}
	return ingredients
}

// creates a list of slices of indices for each ingredient
// variation, splitting on letter enumerated indices
// e.g. 2, 2a, 2b would result in 3 variant lists.
// The maps point to the index of the original ingredient list
func (p *RecipeProcessor) createIndexMap(ingredients []Ingredient) [][]int {
	variantIndicesMap := make(map[int]bool, 0)
	for _, ingredient := range ingredients {
		ind, variant := getIndex(ingredient)
		variantIndicesMap[ind] = variant
	}

	nonVariantIndices := make([]int, 0)
	variantIndices := make([][]int, 0)

	lastIndex := 0
	variantIndex := -1
	for i, ingredient := range ingredients {
		ind, _ := getIndex(ingredient)
		if variantIndicesMap[ind] {
			if ind != lastIndex {
				variantIndex += 1
				variantIndices = append(variantIndices, make([]int, 0))
			}
			variantIndices[variantIndex] = append(variantIndices[variantIndex], i)
			lastIndex = ind
		} else {
			nonVariantIndices = append(nonVariantIndices, i)
		}
	}

	indices := make([][]int, 0)
	indices = append(indices, nonVariantIndices)

	// Loop over variant splits
	for _, variants_type := range variantIndices {
		// reset but make a copy
		baseIndices := make([][]int, len(indices))
		copy(baseIndices, indices)
		indices = make([][]int, 0)

		// Loop over each variant
		for _, variant := range variants_type {
			for _, base := range baseIndices {
				innerBase := make([]int, len(base))
				copy(innerBase, base)
				indices = append(indices, append(innerBase, variant))
			}
		}
	}

	for _, index := range indices {
		sort.Ints(index)
	}

	return indices
}

// returns the numerical index and whether there are options for the ingredient
func getIndex(ingredient Ingredient) (int, bool) {
	ind := 0
	variant := false
	for _, char := range ingredient.Index {
		if char >= '0' && char <= '9' {
			ind = ind*10 + int(char-'0')
		} else if char >= 'a' && char <= 'z' {
			variant = true
			break
		}
	}
	return ind, variant
}
