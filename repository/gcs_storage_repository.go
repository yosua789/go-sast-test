package repository

import (
	"assist-tix/config"
	"bytes"
	"context"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/rs/zerolog/log"
)

type GCSStorageRepositoryImpl struct {
	Client *storage.Client
	Env    *config.EnvironmentVariable
}

type GCSStorageRepository interface {
	WriteFile(fileName string, buffer *bytes.Buffer) error
	ReadFile(fileName string) (reader *storage.Reader, err error)
	CreateSignedUrl(fileName string) (signedUrl string, err error)
}

// WriteFile uploads a file to Google Cloud Storage
func (r *GCSStorageRepositoryImpl) WriteFile(fileName string, buffer *bytes.Buffer) error {
	ctx := context.Background()

	// Create a GCS writer
	writer := r.Client.Bucket(r.Env.Storage.GCS.BucketName).Object(fileName).NewWriter(ctx)
	defer writer.Close()
	// Copy the buffer content to GCS
	_, err := io.Copy(writer, buffer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upload file to GCS")
		return err
	}

	// Close the writer to finalize the upload

	log.Info().Str("Filename", fileName).Msg("File uploaded successfully to GCS")
	return nil
}

// ReadFile downloads a file from Google Cloud Storage
func (r *GCSStorageRepositoryImpl) ReadFile(fileName string) (reader *storage.Reader, err error) {
	ctx := context.Background()

	// Create a GCS reader
	reader, err = r.Client.Bucket(r.Env.Storage.GCS.BucketName).Object(fileName).NewReader(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read file from GCS")
		return nil, err
	}
	log.Info().Str("Filename", fileName).Msg("File read successfully from GCS")
	return reader, nil
}

func (r *GCSStorageRepositoryImpl) CreateSignedUrl(fileName string) (signedUrl string, err error) {

	opts := &storage.SignedURLOptions{
		Method:         "GET",
		Expires:        time.Now().Add(r.Env.Storage.GCS.SignedUrlExpiration),
		GoogleAccessID: r.Env.Storage.GCS.CredentialObj.ClientEmail,
		PrivateKey:     []byte(r.Env.Storage.GCS.CredentialObj.PrivateKey),
		Scheme:         storage.SigningSchemeV4,
	}

	url, err := storage.SignedURL(r.Env.Storage.GCS.BucketName, fileName, opts)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate signed URL")
		return "", err
	}

	signedUrl = url
	return
}

// NewGCSFileRepositoryImpl initializes the GCSFileRepositoryImpl with a GCS client and bucket name
func NewGCSFileRepositoryImpl(client *storage.Client, env *config.EnvironmentVariable) GCSStorageRepository {
	return &GCSStorageRepositoryImpl{
		Client: client,
		Env:    env,
	}
}
