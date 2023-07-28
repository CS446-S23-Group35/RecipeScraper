package cleaner

import (
	"os"

	"github.com/CS446-S23-Group35/RecipeScraper/pkg/recipe"
	"gopkg.in/yaml.v3"
)

type Cleaner interface {
	Clean() error
}

type FileCleaner struct {
	inFile            *os.File
	outFile           *os.File
	BlankImageRemover *BlankRemover
	DedupSorter       *DedupSorter
}

func NewFileCleaner(inFile, outFile *os.File) *FileCleaner {
	return &FileCleaner{
		inFile:            inFile,
		outFile:           outFile,
		BlankImageRemover: NewBlankRemover(),
		DedupSorter:       NewDedupSorter(),
	}
}

func (fc *FileCleaner) Clean() error {
	recipesRaw := make([]recipe.RawRecipe, 0)

	err := yaml.NewDecoder(fc.inFile).Decode(&recipesRaw)
	if err != nil {
		return err
	}

	println("recipesRaw len: ", len(recipesRaw))

	recipesRaw = fc.BlankImageRemover.RemoveBlankImages(recipesRaw)
	recipesRaw = fc.BlankImageRemover.RemoveBlankDescription(recipesRaw)
	recipesRaw = fc.DedupSorter.DedupSort(recipesRaw)

	println("Cleaned recipesRaw len: ", len(recipesRaw))

	return yaml.NewEncoder(fc.outFile).Encode(recipesRaw)
}
