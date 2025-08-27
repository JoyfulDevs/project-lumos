package service

import "context"

type RetrieveParams struct {
	Vectors []float32
	Limit   int32
}

type RetrieveResult struct {
	Score   float32
	Passage []byte
}

type VectorRetriever interface {
	Retrieve(ctx context.Context, params RetrieveParams) ([]RetrieveResult, error)
}

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}
