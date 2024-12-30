package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	internalConfig "github.com/sefazor/ourphotos-backend/internal/config"
)

type CloudflareStorage struct {
	client    *s3.Client
	bucket    string
	accountID string
	publicURL string
}

func NewCloudflareStorage(cfg *internalConfig.Config) (*CloudflareStorage, error) {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2.AccountID),
		}, nil
	})

	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.R2.AccessKeyID,
			cfg.R2.SecretAccessKey,
			"",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &CloudflareStorage{
		client:    s3.NewFromConfig(awsCfg),
		bucket:    cfg.R2.Bucket,
		accountID: cfg.R2.AccountID,
		publicURL: cfg.R2.PublicURL,
	}, nil
}

// Upload dosyayı R2'ye yükler
func (s *CloudflareStorage) Upload(key string, src io.Reader) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   src,
	}

	_, err := s.client.PutObject(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to upload to R2: %v", err)
	}

	return nil
}

// Delete dosyayı R2'den siler
func (s *CloudflareStorage) Delete(key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(context.TODO(), input)
	return err
}
