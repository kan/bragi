package main

import (
	"context"
	"os"
	"strings"

	"github.com/pkg/errors"
	openai "github.com/sashabaranov/go-openai"
)

var openAIClient *openai.Client

func getOpenAIClient(ctx context.Context) (*openai.Client, error) {
	if openAIClient != nil {
		return openAIClient, nil
	}

	client := openai.NewClient(os.Getenv("BRAGI_OPENAI_API_KEY"))

	openAIClient = client
	return client, nil
}

func convertKanjiAI(ctx context.Context, word string) ([]string, error) {
	client, err := getOpenAIClient(ctx)
	if err != nil {
		return []string{}, errors.WithStack(err)
	}

	resp, err := client.CreateChatCompletion(
		ctx,
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
