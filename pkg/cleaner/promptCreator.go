package cleaner

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/gocarina/gocsv"
)

type PromptCreator struct {
	inPromptFiles     []*os.File
	inCompletionFiles []*os.File
	outFile           *os.File
}

type OpenAIPrompt struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
}

type OpenAICompletion struct {
	Index      string `csv:"index"`
	Ingredient string `csv:"basic_ingredient"`
	Amount     string `csv:"amount"`
	Unit       string `csv:"unit"`
	Optional   string `csv:"optional"`
	Notes      string `csv:"notes"`
}

func NewPromptCreator(inPromptFiles, inCompletionFiles []*os.File, outFile *os.File) *PromptCreator {
	return &PromptCreator{
		inPromptFiles:     inPromptFiles,
		inCompletionFiles: inCompletionFiles,
		outFile:           outFile,
	}
}

func (pc *PromptCreator) CreatePrompts() error {
	for i := range pc.inPromptFiles {
		err := pc.createPrompt(pc.inPromptFiles[i], pc.inCompletionFiles[i])
		if err != nil {
			return fmt.Errorf("error parsing file %s:%w", pc.inPromptFiles[i].Name(), err)
		}
	}
	return nil
}

func (pc *PromptCreator) createPrompt(promptFile, completionFile *os.File) error {
	prompt, err := io.ReadAll(promptFile)
	if err != nil {
		return err
	}
	completion, err := io.ReadAll(completionFile)
	if err != nil {
		return err
	}

	completionStructs := make([]OpenAICompletion, 0)
	err = gocsv.UnmarshalBytes(completion, &completionStructs)
	if err != nil {
		return err
	}
	fineTunePrompt := OpenAIPrompt{Prompt: string(prompt), Completion: string(completion)}
	err = json.NewEncoder(pc.outFile).Encode(fineTunePrompt)
	if err != nil {
		return err
	}
	// _, err = pc.outFile.Write([]byte("\n"))
	return nil
}
