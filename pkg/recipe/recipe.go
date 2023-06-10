package recipe

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
	Name        string
	Description string
	Ingredients IngredientList
	Steps       []string

	Metadata RecipeMetadata
}

type RecipeMetadata struct {
	Tags              []string
	MinutesToPrep     int
	MinutesToCook     int
	MinutesTotal      int
	Difficulty        RecipeDifficulty
	Servings          ServingRange
	EstimatedCalories int
	ImageURL          string
	ImageAlt          string
	SourceURL         string

	Dietary RecipeDietaryInformation
}

type ServingRange struct {
	Min int
	Max int
}

type RecipeDietaryInformation struct {
	IsVegetarian bool
	IsVegan      bool
	IsGlutenFree bool
}

type Unit int

const (
	UnitNone Unit = iota
	UnitAbstract
	UnitCup
	UnitPint
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
