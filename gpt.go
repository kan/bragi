package main

import (
	"context"
	"os"
	"strings"

	"github.com/pkg/errors"
	openai "github.com/sashabaranov/go-openai"
)

func convertKanjiAI(ctx context.Context, word string) ([]string, error) {
	client := openai.NewClient(os.Getenv("BRAGI_OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "「" + word + "」を漢字にしてください。固有名詞の場合はカタカナを含んでも構いません。候補が複数ある場合、単語は,で区切ってください。単語のみ返してください",
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
