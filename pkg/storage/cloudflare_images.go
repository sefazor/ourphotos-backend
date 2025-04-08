package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

const (
	VariantPublic    = "public"
	VariantThumbnail = "thumbnail"
)

type CloudflareImages struct {
	accountID   string
	apiToken    string
	baseURL     string
	httpClient  *http.Client
	accountHash string
}

// CloudflareImageResponse represents the response from Cloudflare Images API
type CloudflareImageResponse struct {
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	Result struct {
		ID       string   `json:"id"`
		Variants []string `json:"variants"`
	} `json:"result"`
}

func NewCloudflareImages(accountID, token, accountHash string) *CloudflareImages {
	// Optimize edilmiş HTTP istemcisi oluştur
	client := &http.Client{
		Timeout: 5 * time.Minute, // Büyük dosya yüklemeleri için daha uzun timeout
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     100,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
			ForceAttemptHTTP2:   true,
		},
	}

	return &CloudflareImages{
		accountID:   accountID,
		apiToken:    token,
		baseURL:     "https://api.cloudflare.com/client/v4/accounts/%s/images/v1",
		httpClient:  client,
		accountHash: accountHash,
	}
}

// Upload dosyayı Cloudflare Images'a yükler ve image ID ve variantları döndürür
func (c *CloudflareImages) Upload(reader io.Reader) (string, []string, error) {
	return c.UploadWithFilename(reader, "image.jpg")
}

// UploadWithFilename orijinal dosya adıyla birlikte yükleme yapar
func (c *CloudflareImages) UploadWithFilename(reader io.Reader, filename string) (string, []string, error) {
	fmt.Printf("Cloudflare Images Upload başlatılıyor... Dosya adı: %s\n", filename)

	// Önce tüm içeriği bir buffer'a oku
	var buf bytes.Buffer
	fileSize, err := io.Copy(&buf, reader)
	if err != nil {
		fmt.Printf("Dosya içeriği buffer'a kopyalanamadı: %v\n", err)
		return "", nil, fmt.Errorf("failed to copy file content to buffer: %w", err)
	}

	if fileSize == 0 {
		fmt.Printf("HATA: Dosya boyutu 0\n")
		return "", nil, fmt.Errorf("empty file, size is 0 bytes")
	}

	fmt.Printf("Dosya içeriği buffer'a kopyalandı, boyut: %d bytes\n", fileSize)

	// Dosya içeriğini koruyoruz
	fileBytes := buf.Bytes()

	// Multipart form hazırlama fonksiyonu
	createForm := func() (*bytes.Buffer, string, error) {
		formBuf := &bytes.Buffer{}
		writer := multipart.NewWriter(formBuf)

		// Form alanını oluştur - orijinal dosya adını kullan
		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create form file: %w", err)
		}

		// Dosya içeriğini yaz
		if _, err := io.Copy(part, bytes.NewReader(fileBytes)); err != nil {
			return nil, "", fmt.Errorf("failed to copy file: %w", err)
		}

		// Diğer form alanlarını ekle
		if err := writer.WriteField("requireSignedURLs", "false"); err != nil {
			return nil, "", fmt.Errorf("failed to add form field: %w", err)
		}

		// Form'u kapat
		if err := writer.Close(); err != nil {
			return nil, "", fmt.Errorf("failed to close writer: %w", err)
		}

		return formBuf, writer.FormDataContentType(), nil
	}

	// Form oluştur
	formBuf, contentType, err := createForm()
	if err != nil {
		return "", nil, fmt.Errorf("failed to create multipart form: %w", err)
	}

	// HTTP isteği için URL hazırla
	cloudflareURL := fmt.Sprintf(c.baseURL, c.accountID)

	// HTTP isteği hazırla
	req, err := http.NewRequest("POST", cloudflareURL, formBuf)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	// GetBody fonksiyonunu ekle - HTTP/2 retry için gerekli
	req.GetBody = func() (io.ReadCloser, error) {
		newForm, _, err := createForm()
		if err != nil {
			return nil, err
		}
		return io.NopCloser(newForm), nil
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	// İsteği gönder
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		return "", nil, fmt.Errorf("cloudflare returned non-OK status: %d, response: %s", resp.StatusCode, bodyString)
	}

	// Decode response
	var response CloudflareImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return "", nil, fmt.Errorf("cloudflare returned error: %v", response.Errors)
	}

	// Create variant URLs
	variantURLs := []string{
		c.GetPublicURL(response.Result.ID),
		c.GetThumbnailURL(response.Result.ID),
	}

	return response.Result.ID, variantURLs, nil
}

func (c *CloudflareImages) Delete(imageID string) error {
	url := fmt.Sprintf(c.baseURL+"/%s", c.accountID, imageID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete image: %d", resp.StatusCode)
	}

	return nil
}

func (c *CloudflareImages) GetPublicURL(imageID string) string {
	return fmt.Sprintf("https://imagedelivery.net/%s/%s/%s", c.accountHash, imageID, VariantPublic)
}

func (c *CloudflareImages) GetThumbnailURL(imageID string) string {
	return fmt.Sprintf("https://imagedelivery.net/%s/%s/%s", c.accountHash, imageID, VariantThumbnail)
}
