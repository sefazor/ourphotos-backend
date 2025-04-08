package storage

import "io"

type StorageService interface {
	Upload(key string, reader io.Reader) error
	Delete(key string) error
}

type ImageService interface {
	Upload(reader io.Reader) (string, []string, error) // returns imageID, variants, error
	Delete(imageID string) error
	GetPublicURL(imageID string) string
	GetThumbnailURL(imageID string) string
}
