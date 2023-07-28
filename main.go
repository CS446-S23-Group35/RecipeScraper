package main

import (
	"os"

	"github.com/CS446-S23-Group35/RecipeScraper/pkg/processor"
	"github.com/CS446-S23-Group35/RecipeScraper/pkg/recipe"
	"gopkg.in/yaml.v3"
)

func main() {
	//////////////////////
	// Scraping
	//////////////////////

	// cfg := scraper.Config{
	// 	StartLink:  "https://www.foodnetwork.com/recipes/recipes-a-z/l/p/12",
	// 	SourceType: "foodnetwork",
	// 	OutputPath: "recipes.yaml",
	// 	OnlyLinks:  true,
	// }

	// scraper := scraper.NewScraper(cfg)
	// err := scraper.Scrape(context.Background())
	// if err != nil {
	// 	panic(err)
	// }

	// err := scraper.ScrapeFromLinksFile(context.Background(), "links.tmp")
	// if err != nil {
	// 	panic(err)
	// }

	//////////////////////
	// Cleaning
	//////////////////////

	// inFile, err := os.Open("recipes/recipesNoImages.yaml")
	// if err != nil {
	// 	panic(err)
	// }
	// defer inFile.Close()

	// outFile, err := os.Create("recipes/recipes.yaml")
	// if err != nil {
	// 	panic(err)
	// }
	// defer outFile.Close()

	// c := cleaner.NewFileCleaner(inFile, outFile)
	// if err := c.Clean(); err != nil {
	// 	fmt.Println("error deduping: " + err.Error())
	// }

	//////////////////////
	// Prompting
	//////////////////////

	// 	openai := prompter.NewOpenAIPrompter()
	// 	ings := `1/2 cup brandy
	// 6 nectarines, skin left on, halved, pit removed
	// 1 tablespoon sugar, to flambe
	// 1/2 cup milk
	// 1/2 cup heavy cream
	// 1 vanilla pod or 3 teaspoons vanilla extract or 1 tbps pure vanilla extract
	// 1 tablespoon lemon liqueur or 3 teaspoons lemon extract (recommended: Limoncello)
	// 1/4 cup sugar, for lemon cream
	// 4 egg yolks
	// 1 tablespoon finely grated lemon zest
	// 1 pint blackberries`
	// ings := `- Cooking spray
	// - 2 graham crackers
	// - 1 tablespoon unsalted butter, melted
	// - 1 teaspoon sugar
	// - 3 ounces cream cheese, softened
	// - 3 tablespoons sour cream
	// - 3 tablespoons sugar
	// - 1 large egg
	// - 2 teaspoons all-purpose flour
	// - 1/2 teaspoon lemon zest
	// - 1/2 teaspoon pure vanilla extract
	// - Pinch of salt
	// - 2 large strawberries, finely chopped
	// - 1 teaspoon strawberry jam`

	// req := prompter.NewParseIngredientsRequest(ings)

	// resp, err := openai.MakeRequest(req)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(resp.Choices[0].Text)

	////////////////////////////////////////
	// Processing - First step, ingredients
	////////////////////////////////////////

	// file, err := os.Open("recipes/recipes.yaml")
	// if err != nil {
	// 	panic(err)
	// }
	// defer file.Close()

	// recipes := make([]*recipe.RawRecipe, 0)
	// err = yaml.NewDecoder(file).Decode(&recipes)
	// if err != nil {
	// 	panic(err)
	// }

	// otherFile, err := os.Open("recipes/referenceRecipes.yaml")
	// if err != nil {
	// 	panic(err)
	// }
	// defer otherFile.Close()
	// filterBase := make([]*recipe.RawRecipe, 0)
	// err = yaml.NewDecoder(otherFile).Decode(&filterBase)
	// if err != nil {
	// 	panic(err)
	// }
	// urlSet := make(map[string]bool)
	// for _, r := range filterBase {
	// 	urlSet[r.Metadata.SourceURL] = true
	// }

	// filtered := make([]*recipe.RawRecipe, 0, len(recipes))
	// for _, r := range recipes {
	// 	if !urlSet[r.Metadata.SourceURL] {
	// 		filtered = append(filtered, r)
	// 	}
	// }

	// processor := processor.NewRecipeProcessor()
	// defer processor.Close()

	// err = processor.ProcessRawRecipes(filtered, len(filtered), 3)
	// if err != nil {
	// 	panic(err)
	// }

	////////////////////////////////////////
	// Processing - Second step, attributes
	////////////////////////////////////////

	file, err := os.Open("recipes/Backup.yaml")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	filterBase := make([]*recipe.Recipe, 0)
	err = yaml.NewDecoder(file).Decode(&filterBase)
	if err != nil {
		panic(err)
	}

	groupedByUrl := make(map[string][]*recipe.Recipe)
	for _, r := range filterBase {
		groupedByUrl[r.Metadata.SourceURL] = append(groupedByUrl[r.Metadata.SourceURL], r)
	}

	processor := processor.NewRecipeProcessor()
	defer processor.Close()
	err = processor.ProcessRecipeAttributes(groupedByUrl, len(groupedByUrl), 2)
	if err != nil {
		panic(err)
	}
	// for _, v := range groupedByUrl {
	// 	res, err := processor.ProcessAttributes(v)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	} else {
	// 		for _, r := range res {
	// 			fmt.Println(r.Name)
	// 			fmt.Println(r.Metadata.Dietary)
	// 		}
	// 	}
	// }

}
