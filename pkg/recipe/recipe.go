package recipe

import "fmt"

type RecipeDifficulty int

const (
	Easy   RecipeDifficulty = 1
	Medium RecipeDifficulty = 3
	Hard   RecipeDifficulty = 5
)

type IngredientItem struct {
	Name     string
	Amount   Amount
	Optional bool
	Notes    string
}

type IngredientList []IngredientItem

type Recipe struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Ingredients IngredientList `yaml:"ingredients"`
	Steps       []string       `yaml:"steps"`

	Metadata RecipeMetadata `yaml:"metadata"`
}

type RecipeMetadata struct {
	Tags              []string         `yaml:"tags"`
	MinutesToPrep     int              `yaml:"minutes_to_prep"`
	MinutesToCook     int              `yaml:"minutes_to_cook"`
	MinutesTotal      int              `yaml:"minutes_total"`
	Difficulty        RecipeDifficulty `yaml:"difficulty"`
	Servings          ServingRange     `yaml:"servings"`
	EstimatedCalories int              `yaml:"estimated_calories"`
	ImageURL          string           `yaml:"image_url"`
	ImageAlt          string           `yaml:"image_alt"`
	SourceURL         string           `yaml:"source_url"`

	Dietary RecipeDietaryInformation `yaml:"dietary"`
}

type ServingRange struct {
	Min         int    `yaml:"min"`
	Max         int    `yaml:"max"`
	Alternative string `yaml:"alternative"`
}

type RecipeDietaryInformation struct {
	IsVegetarian    bool `yaml:"is_vegetarian"`
	IsVegan         bool `yaml:"is_vegan"`
	IsGlutenFree    bool `yaml:"is_gluten_free"`
	IsDairyFree     bool `yaml:"is_dairy_free"`
	IsNutFree       bool `yaml:"is_nut_free"`
	IsShellfishFree bool `yaml:"is_shellfish_free"`
	IsEggFree       bool `yaml:"is_egg_free"`
	IsSoyFree       bool `yaml:"is_soy_free"`
	IsFishFree      bool `yaml:"is_fish_free"`
	IsPorkFree      bool `yaml:"is_pork_free"`
	IsRedMeatFree   bool `yaml:"is_red_meat_free"`
	IsAlcoholFree   bool `yaml:"is_alcohol_free"`
	IsKosher        bool `yaml:"is_kosher"`
	IsHalal         bool `yaml:"is_halal"`
}

type Unit int

func UnitFromStr(str string) Unit {
	unitTable := map[string]Unit{
		"":             UnitNone,
		"-":            UnitNone,
		"tbsp":         UnitTbsp,
		"tablespoon":   UnitTbsp,
		"tablespoons":  UnitTbsp,
		"tsp":          UnitTsp,
		"teaspoon":     UnitTsp,
		"teaspoons":    UnitTsp,
		"cp":           UnitCup,
		"cup":          UnitCup,
		"cups":         UnitCup,
		"pt":           UnitPint,
		"pint":         UnitPint,
		"pints":        UnitPint,
		"fl oz":        UnitFlOz,
		"floz":         UnitFlOz,
		"fluid ounce":  UnitFlOz,
		"fluid ounces": UnitFlOz,
		"oz":           UnitOz,
		"ounce":        UnitOz,
		"ounces":       UnitOz,
		"lb":           UnitLb,
		"lbs":          UnitLb,
		"pound":        UnitLb,
		"pounds":       UnitLb,
		"g":            UnitGram,
		"gram":         UnitGram,
		"grams":        UnitGram,
		"kg":           UnitKg,
		"kilogram":     UnitKg,
		"kilograms":    UnitKg,
		"quanity":      UnitQuanity,
		"quanities":    UnitQuanity,
		"qty":          UnitQuanity,
		"qt":           UnitQuart,
		"quart":        UnitQuart,
		"quarts":       UnitQuart,
		"gal":          UnitGallon,
		"gallon":       UnitGallon,
		"gallons":      UnitGallon,
	}

	if unit, ok := unitTable[str]; ok {
		return unit
	} else {
		return UnitAbstract
	}
}

// Parses a fraction string of the form w n/d into a float64.
// returns the float64 and a bool indicating if it is a fraction
// return -1 if not a number
func ParseFraction(str string) (float64, bool) {
	w, n, d := 0, 0, 0

	_, err := fmt.Sscanf(str, "%d %d/%d", &w, &n, &d)
	if err == nil {
		return float64(w) + float64(n)/float64(d), true
	}

	_, err = fmt.Sscanf(str, "%d/%d", &n, &d)
	if err == nil {
		return float64(n) / float64(d), true
	}

	_, err = fmt.Sscanf(str, "%d", &w)
	if err == nil {
		return float64(w), false
	}
	return -1, false
}

const (
	UnitNone Unit = iota
	UnitAbstract
	UnitCup
	UnitPint
	UnitQuart
	UnitGallon
	UnitFlOz
	UnitTsp
	UnitTbsp
	UnitOz
	UnitLb
	UnitGram
	UnitKg
	UnitQuanity
	UnitFraction
)

type Amount struct {
	Type     Unit
	TypeName string
	Value    float64
}
