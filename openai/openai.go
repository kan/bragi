package openai

import (
	"context"
	"os"
	"strings"

	"github.com/pkg/errors"
	openai "github.com/sashabaranov/go-openai"
)

type OpenAIDict struct {
	client *openai.Client
}

func (d *OpenAIDict) Convert(word string) ([]string, error) {
	if len(word) < 2 {
		return []string{}, nil
	}

	resp, err := d.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "以下の文章をかな漢字変換してください。候補が複数ある場合、,で区切ってください\n\n" + word,
				},
			},
		},
	)
	if err != nil {
		return []string{}, errors.WithStack(err)
	}

	content := resp.Choices[0].Message.Content
	return strings.Split(content, ","), nil
}

func NewOpenAIDict() *OpenAIDict {
	client := openai.NewClient(os.Getenv("BRAGI_OPENAI_API_KEY"))
	return &OpenAIDict{client: client}
}
