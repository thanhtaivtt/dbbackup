package storage

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/thanhtaivtt/dbbackup/internal/config"
)

type R2Storage struct {
	client *s3.Client
	bucket string
}

func NewR2(cfg config.R2Config) (*R2Storage, error) {
	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)),
		Region:       "auto",
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.AccessKeySecret, ""),
	})

	return &R2Storage{client: client, bucket: cfg.Bucket}, nil
}

func (s *R2Storage) Name() string { return "r2" }

func (s *R2Storage) Upload(ctx context.Context, key string, reader io.Reader) error {
	// Buffer to temp file for seekable upload (S3 SDK needs rewind on retry)
	tmp, err := os.CreateTemp("", "dbbackup-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := io.Copy(tmp, reader); err != nil {
		return fmt.Errorf("buffering upload data: %w", err)
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seeking temp file: %w", err)
	}

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
		Body:   tmp,
	})
	if err != nil {
		return fmt.Errorf("r2 upload %s: %w", key, err)
	}
	return nil
}

func (s *R2Storage) List(ctx context.Context, prefix string) ([]Object, error) {
	var objects []Object
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: &s.bucket,
		Prefix: &prefix,
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("r2 list: %w", err)
		}
		for _, obj := range page.Contents {
			objects = append(objects, Object{
				Key:          aws.ToString(obj.Key),
				Size:         aws.ToInt64(obj.Size),
				LastModified: aws.ToTime(obj.LastModified),
			})
		}
	}
	return objects, nil
}

func (s *R2Storage) Delete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	deleteObjects := make([]types.ObjectIdentifier, len(keys))
	for i, key := range keys {
		deleteObjects[i] = types.ObjectIdentifier{Key: aws.String(key)}
	}

	_, err := s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: &s.bucket,
		Delete: &types.Delete{Objects: deleteObjects},
	})
	if err != nil {
		return fmt.Errorf("r2 delete: %w", err)
	}
	return nil
}
