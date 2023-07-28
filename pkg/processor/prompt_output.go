package processor

type Ingredient struct {
	Index      string `csv:"index"`
	Ingredient string `csv:"basic_ingredient"`
	Amount     string `csv:"amount"`
	Unit       string `csv:"unit"`
	Optional   string `csv:"optional"`
	Notes      string `csv:"notes"`
}
