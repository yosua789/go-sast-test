package storage

import (
	"assist-tix/config"
	"context"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

func NewGCSClient(env *config.EnvironmentVariable) (*storage.Client, error) {
	ctx := context.Background()

	bytes, err := env.Storage.GCS.CredentialObj.ToBytes()
	if err != nil {
		return nil, err
	}

	optionCred := option.WithCredentialsJSON(bytes)

	client, err := storage.NewClient(ctx, optionCred)
	if err != nil {
		return nil, err
	}
	return client, nil
}
