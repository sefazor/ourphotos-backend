package config

import (
	"fmt"
	"os"
)

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	PublicURL       string
}

type Config struct {
	R2               R2Config
	CloudflareImages struct {
		AccountID string
		Token     string
		Hash      string // Images CDN URL'leri için hash değeri
	}
}

func LoadConfig() *Config {
	cfg := &Config{}

	// R2 config
	cfg.R2.AccountID = os.Getenv("R2_ACCOUNT_ID")
	cfg.R2.AccessKeyID = os.Getenv("R2_ACCESS_KEY_ID")
	cfg.R2.SecretAccessKey = os.Getenv("R2_SECRET_ACCESS_KEY")
	cfg.R2.Bucket = os.Getenv("R2_BUCKET")
	cfg.R2.PublicURL = os.Getenv("R2_PUBLIC_URL")

	// Cloudflare Images config
	cfg.CloudflareImages.AccountID = os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	cfg.CloudflareImages.Token = os.Getenv("CLOUDFLARE_IMAGES_TOKEN")
	cfg.CloudflareImages.Hash = os.Getenv("CLOUDFLARE_IMAGES_HASH")

	// Debug için
	fmt.Printf("Debug - Loading Cloudflare config: AccountID=%s, TokenLength=%d, Hash=%s\n",
		cfg.CloudflareImages.AccountID, len(cfg.CloudflareImages.Token), cfg.CloudflareImages.Hash)

	return cfg
}
