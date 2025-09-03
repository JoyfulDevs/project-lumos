package adapter

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/joyfuldevs/project-lumos/cmd/dense-retrieval/app/service"
)

var _ service.Embedder = (*OpenAIClient)(nil)

type OpenAIClient struct {
	client *openai.Client
}

func NewOpenAIClient(baseURL string) *OpenAIClient {
	client := openai.NewClient(
		option.WithBaseURL(baseURL),
	)

	return &OpenAIClient{client: &client}
}

func (o *OpenAIClient) Embed(ctx context.Context, text string) ([]float32, error) {
	resp, err := o.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(text),
		},
		EncodingFormat: openai.EmbeddingNewParamsEncodingFormatFloat,
	})
	if err != nil {
		return nil, err
	}
	vectors := make([]float32, len(resp.Data[0].Embedding))
	for i, v := range resp.Data[0].Embedding {
		vectors[i] = float32(v)
	}
	return vectors, nil
}
