package service

type Service struct {
	VectorRetriever VectorRetriever
	Embedder        Embedder
}

func NewService(v VectorRetriever, e Embedder) *Service {
	return &Service{
		VectorRetriever: v,
		Embedder:        e,
	}
}
