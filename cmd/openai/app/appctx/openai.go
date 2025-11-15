package appctx

import "context"

type openAIAPIOptionType int

const (
	openAIAPIOptionKey openAIAPIOptionType = iota
	openaiAPIOptionURL
)

func WithOpenAIAPIKey(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, openAIAPIOptionKey, value)
}

func OpenAIAPIKeyFrom(ctx context.Context) string {
	info, _ := ctx.Value(openAIAPIOptionKey).(string)
	return info
}

func WithOpenAIAPIURL(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, openaiAPIOptionURL, value)
}

func OpenAIAPIURLFrom(ctx context.Context) string {
	info, _ := ctx.Value(openaiAPIOptionURL).(string)
	return info
}
