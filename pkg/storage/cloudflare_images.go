package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type CloudflareImages struct {
	accountID   string
	apiToken    string
	baseURL     string
	accountHash string
}

func NewCloudflareImages() *CloudflareImages {
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	apiToken := os.Getenv("CLOUDFLARE_IMAGES_TOKEN")
	accountHash := os.Getenv("CLOUDFLARE_IMAGES_HASH")

	baseURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/images/v1", accountID)

	return &CloudflareImages{
		accountID:   accountID,
		apiToken:    apiToken,
		baseURL:     baseURL,
		accountHash: accountHash,
	}
}

func (c *CloudflareImages) UploadImage(file *multipart.FileHeader) (string, []string, error) {
	src, err := file.Open()
	if err != nil {
		return "", nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", file.Filename)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create form file: %v", err)
	}

	if _, err = io.Copy(part, src); err != nil {
		return "", nil, fmt.Errorf("failed to copy file: %v", err)
	}

	if err := writer.Close(); err != nil {
		return "", nil, fmt.Errorf("failed to close writer: %v", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, body)
	if err != nil {
		return "", nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Result struct {
			ID string `json:"id"`
		} `json:"result"`
		Success bool `json:"success"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if !result.Success {
		return "", nil, fmt.Errorf("failed to upload image")
	}

	variants := []string{
		fmt.Sprintf("https://imagedelivery.net/%s/%s/public", c.accountHash, result.Result.ID),
		fmt.Sprintf("https://imagedelivery.net/%s/%s/thumbnail", c.accountHash, result.Result.ID),
		fmt.Sprintf("https://imagedelivery.net/%s/%s/medium", c.accountHash, result.Result.ID),
	}

	return result.Result.ID, variants, nil
}

func (c *CloudflareImages) Delete(imageID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", c.baseURL, imageID), nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete image: %d", resp.StatusCode)
	}

	return nil
}
