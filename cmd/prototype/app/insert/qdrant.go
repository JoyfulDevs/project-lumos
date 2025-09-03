package app

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"slices"
	"strconv"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"

	"github.com/joyfuldevs/project-lumos/cmd/prototype/app"
)

func Insert(
	ctx context.Context,
	address string,
	name string,
	dimension uint64,
	embeddings []app.Embedding,
) error {
	client, err := createClient(address)
	if err != nil {
		return err
	}
	defer func() {
		if err := client.Close(); err != nil {
			slog.Warn("failed to close Qdrant client", slog.String("error", err.Error()))
		}
	}()

	if err := createCollection(ctx, client, name, dimension); err != nil {
		return err
	}

	points := make([]*qdrant.PointStruct, 0, len(embeddings))
	for _, embedding := range embeddings {
		points = append(points, &qdrant.PointStruct{
			Id:      qdrant.NewIDUUID(uuid.NewString()),
			Payload: qdrant.NewValueMap(embedding.Payload),
			Vectors: qdrant.NewVectors(embedding.Vectors...),
		})
	}

	_, err = client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: name,
		Points:         points,
	})

	if err != nil {
		return err
	}

	return nil
}

func createClient(address string) (*qdrant.Client, error) {
	host, portString, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		return nil, err
	}
	return qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})
}

func createCollection(ctx context.Context, client *qdrant.Client, name string, dimension uint64) error {
	names, err := client.ListCollections(ctx)
	if err != nil {
		return err
	}

	if slices.Contains(names, name) {
		return nil
	}

	return client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: name,
		VectorsConfig: qdrant.NewVectorsConfig(
			&qdrant.VectorParams{
				Size:     dimension,
				Distance: qdrant.Distance_Cosine,
			},
		),
	})
}
