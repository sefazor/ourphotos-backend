package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type CloudflareStorage struct {
	client    *s3.Client
	bucket    string
	accountID string
	publicURL string
}

func NewCloudflareStorage() (*CloudflareStorage, error) {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKeyID := os.Getenv("R2_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucket := os.Getenv("R2_BUCKET")
	publicURL := os.Getenv("R2_PUBLIC_URL")

	if accountID == "" || accessKeyID == "" || secretAccessKey == "" || bucket == "" || publicURL == "" {
		return nil, fmt.Errorf("missing required CloudFlare R2 configuration")
	}

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	client := s3.NewFromConfig(cfg)

	return &CloudflareStorage{
		client:    client,
		bucket:    bucket,
		accountID: accountID,
		publicURL: publicURL,
	}, nil
}

// Upload dosyayı R2'ye yükler
func (s *CloudflareStorage) Upload(key string, src multipart.File) error {
	data, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	}

	_, err = s.client.PutObject(context.TODO(), input)
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
