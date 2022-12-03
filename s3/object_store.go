package s3

import (
	"context"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ObjectStore struct {
	Client *s3.Client
	log    *log.Logger
}

type NewObjectStoreOptions struct {
	Config    aws.Config
	Log       *log.Logger
	PathStyle bool
}

// NewObjectStore with the given options.
// If no logger is provided, logs are discarded.
func NewObjectStore(opts NewObjectStoreOptions) *ObjectStore {
	if opts.Log == nil {
		opts.Log = log.New(io.Discard, "", 0)
	}

	client := s3.NewFromConfig(opts.Config, func(o *s3.Options) {
		o.UsePathStyle = opts.PathStyle
	})

	return &ObjectStore{
		Client: client,
		log:    opts.Log,
	}
}

// Put an object in the bucket under key.
func (b *ObjectStore) Put(ctx context.Context, bucket, key, contentType string, body io.Reader) error {
	_, err := b.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        body,
		ContentType: &contentType,
	})
	return err
}

// Get an object from the bucket under key.
// If there is nothing there, returns nil and no error.
func (b *ObjectStore) Get(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	getObjectOutput, err := b.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if getObjectOutput == nil {
		return nil, nil
	}
	return getObjectOutput.Body, err
}

// Delete an object from the bucket under key.
// Deleting where nothing exists does nothing and returns no error.
func (b *ObjectStore) Delete(ctx context.Context, bucket, key string) error {
	_, err := b.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	return err
}
