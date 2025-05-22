package inference

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/JerkyTreats/PHITE/converter/internal/models"
	"github.com/JerkyTreats/PHITE/converter/pkg/logger"
	"github.com/openai/openai-go"
)

type Inference struct{}

func NewInference() *Inference {
	return &Inference{}
}

func (i *Inference) getSystemPrompt() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("unable to get current file path")
	}
	dir := filepath.Dir(filename)
	promptPath := filepath.Join(dir, "system_prompt.md")
	data, err := os.ReadFile(promptPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (i *Inference) postGroupingInference(group *models.Grouping) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		logger.Fatal(fmt.Errorf("OPENAI_API_KEY not set in environment"), "OPENAI_API_KEY not set in environment")
	}

	client := openai.NewClient()
	ctx := context.Background()

	// Build user message with Grouping details
	userMessage := `Given the following Grouping of SNPs, generate a structured, layered genetic insight response.

Please return: Overview paragraph Key Takeaways (Markdown bullets) Mitigation Strategies (as JSON) Gene-Specific Insights (as JSON)
`
	userMessage += group.ToString()

	systemPrompt, err := i.getSystemPrompt()
	if err != nil {
		logger.Fatal(err, err.Error())
	}

	// Prepare OpenAI chat completion request using openai-go SDK
	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4o,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userMessage),
		},
	}
	stream := client.Chat.Completions.NewStreaming(ctx, params)
	acc := openai.ChatCompletionAccumulator{}
	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)
	}
	if err := stream.Err(); err != nil {
		return "", err
	}
	if len(acc.Choices) == 0 || acc.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("no content returned from OpenAI")
	}
	return acc.Choices[0].Message.Content, nil
}
