package storage

import (
	"bytes"
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
	fmt.Printf("R2 Storage - Upload başlatılıyor: %s\n", key)

	// Önce, eğer bu bir dosya ise ve boyutunu biliyorsak, doğrudan kullanabiliriz
	if readerWithSize, ok := src.(io.ReadSeeker); ok {
		fmt.Printf("R2 Storage - ReadSeeker mevcut, dosya boyutu ölçülüyor\n")

		// Dosya boyutunu ölç
		currentPos, err := readerWithSize.Seek(0, io.SeekCurrent)
		if err != nil {
			fmt.Printf("R2 Storage - Mevcut pozisyon alınamadı: %v\n", err)
			return fmt.Errorf("failed to get current position: %w", err)
		}

		// Sonuna git, boyutu ölç
		size, err := readerWithSize.Seek(0, io.SeekEnd)
		if err != nil {
			fmt.Printf("R2 Storage - Dosya sonuna gidilemedi: %v\n", err)
			return fmt.Errorf("failed to seek to end: %w", err)
		}
		fmt.Printf("R2 Storage - Dosya boyutu: %d bytes\n", size)

		// Başlangıç pozisyonuna geri dön
		_, err = readerWithSize.Seek(currentPos, io.SeekStart)
		if err != nil {
			fmt.Printf("R2 Storage - Başlangıç pozisyonuna dönülemedi: %v\n", err)
			return fmt.Errorf("failed to seek back to start: %w", err)
		}

		input := &s3.PutObjectInput{
			Bucket:        aws.String(s.bucket),
			Key:           aws.String(key),
			Body:          src,
			ContentLength: aws.Int64(size),
		}

		fmt.Printf("R2 Storage - Dosya R2'ye yükleniyor (ReadSeeker): %s, bucket: %s\n", key, s.bucket)
		_, err = s.client.PutObject(context.TODO(), input)
		if err != nil {
			fmt.Printf("R2 Storage - Yükleme hatası: %v\n", err)
			return fmt.Errorf("failed to upload to R2: %w", err)
		}

		fmt.Printf("R2 Storage - Dosya başarıyla yüklendi: %s\n", key)
		return nil
	}

	fmt.Printf("R2 Storage - ReadSeeker mevcut değil, dosya tamamen okunuyor\n")
	// Aksi takdirde, içeriği okumalıyız - ama çok büyük dosyalar için bellek kullanımını azaltmak için
	// daha gelişmiş bir yaklaşım gerekebilir
	buf, err := io.ReadAll(src)
	if err != nil {
		fmt.Printf("R2 Storage - Dosya içeriği okunamadı: %v\n", err)
		return fmt.Errorf("failed to read file content: %w", err)
	}

	fmt.Printf("R2 Storage - Dosya tamamen okundu, boyut: %d bytes\n", len(buf))
	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(buf),
		ContentLength: aws.Int64(int64(len(buf))),
	}

	fmt.Printf("R2 Storage - Dosya R2'ye yükleniyor (buffer): %s, bucket: %s\n", key, s.bucket)
	_, err = s.client.PutObject(context.TODO(), input)
	if err != nil {
		fmt.Printf("R2 Storage - Yükleme hatası: %v\n", err)
		return fmt.Errorf("failed to upload to R2: %w", err)
	}

	fmt.Printf("R2 Storage - Dosya başarıyla yüklendi: %s\n", key)
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
