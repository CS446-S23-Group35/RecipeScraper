package prompter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type OpenAIPrompter struct {
	token string
}

type OpenAiResponse struct {
	Object  string `json:"object"`
	Created int    `json:"created"`
	Choices []struct {
		Text  string `json:"text"`
		Index int    `json:"index"`
	}
	Usage struct {
		PromtTokens      int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func NewOpenAIPrompter() *OpenAIPrompter {
	token, err := os.ReadFile("secret/openai.token")
	if err != nil {
		panic(fmt.Errorf("failed to open openai token file: %w", err))
	}

	return &OpenAIPrompter{
		token: string(token),
	}
}

func (p *OpenAIPrompter) MakeRequest(req OpenAIRequest) (*OpenAiResponse, error) {
	reqBody, err := req.MakeBody()
	if err != nil {
		return nil, fmt.Errorf("failed to make request body: %w", err)
	}

	openAiReq, err := http.NewRequest(http.MethodPost, req.URL(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	openAiReq.Header.Add("Authorization", "Bearer "+p.token)
	openAiReq.Header.Add("Content-Type", "application/json")

	bOff := newBackoff()
	for {
		resp, err := http.DefaultClient.Do(openAiReq)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %w", err)
		}

		defer resp.Body.Close()
		if resp.StatusCode == http.StatusTooManyRequests {
			timeWait := bOff.NextBackOff()
			fmt.Printf("Too many requests, waiting %f seconds\n", timeWait.Seconds())
			time.Sleep(timeWait)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("bad status code: %s", resp.Status)
		}

		respParsed := &OpenAiResponse{}
		err = json.NewDecoder(resp.Body).Decode(respParsed)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}

		return respParsed, nil
	}
}

func newBackoff() backoff.BackOff {
	bOff := backoff.NewExponentialBackOff()
	bOff.MaxElapsedTime = time.Minute
	bOff.Multiplier = 1.05
	bOff.InitialInterval = 2 * time.Second
	return bOff
}

// text-babbage-001, babbage:2020-05-03, text-babbage:001, babbage
