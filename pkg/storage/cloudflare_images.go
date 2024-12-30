package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type CloudflareImages struct {
	accountID string
	apiToken  string
	baseURL   string
}

const (
	VariantPublic    = "public"    // Orijinal boyut
	VariantThumbnail = "thumbnail" // 100x100 thumbnail
	VariantMedium    = "medium"    // 300x300
	VariantLarge     = "large"     // 700x700
)

func NewCloudflareImages(accountID, token string) *CloudflareImages {
	fmt.Printf("Debug - Initializing Cloudflare Images with Account ID: %s\n", accountID)
	fmt.Printf("Debug - API Token length: %d\n", len(token))

	return &CloudflareImages{
		accountID: accountID,
		apiToken:  token,
		baseURL:   "https://api.cloudflare.com/client/v4",
	}
}

func (c *CloudflareImages) Upload(reader io.Reader) (string, []string, error) {
	fileBytes, err := io.ReadAll(reader)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read file: %w", err)
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/images/v1", c.accountID)
	fmt.Printf("Debug - Making request to: %s\n", url)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "image.jpg")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(fileBytes); err != nil {
		return "", nil, fmt.Errorf("failed to write file bytes: %w", err)
	}

	writer.WriteField("requireSignedURLs", "false")

	if err := writer.Close(); err != nil {
		return "", nil, fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	fmt.Printf("Debug - Request Headers: %+v\n", req.Header)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("Debug - Response Status: %d\n", resp.StatusCode)
	fmt.Printf("Debug - Response Body: %s\n", string(respBody))

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("cloudflare upload failed with status %d: %s",
			resp.StatusCode, string(respBody))
	}

	var result struct {
		Success bool `json:"success"`
		Result  struct {
			ID       string   `json:"id"`
			Variants []string `json:"variants"`
		} `json:"result"`
		Errors []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&result); err != nil {
		return "", nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		if len(result.Errors) > 0 {
			return "", nil, fmt.Errorf("cloudflare error: %s", result.Errors[0].Message)
		}
		return "", nil, fmt.Errorf("cloudflare upload failed without specific error")
	}

	return result.Result.ID, result.Result.Variants, nil
}

func (c *CloudflareImages) Delete(imageID string) error {
	url := fmt.Sprintf("%s/accounts/%s/images/v1/%s", c.baseURL, c.accountID, imageID)

	req, err := http.NewRequest("DELETE", url, nil)
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

func (c *CloudflareImages) GetPublicURL(imageID string) string {
	return fmt.Sprintf("https://imagedelivery.net/%s/%s/public", c.accountID, imageID)
}

func (c *CloudflareImages) GetVariantURL(imageID string, variant string) string {
	return fmt.Sprintf("https://imagedelivery.net/%s/%s/%s", c.accountID, imageID, variant)
}

func (c *CloudflareImages) GetThumbnailURL(imageID string) string {
	return c.GetVariantURL(imageID, VariantThumbnail)
}

func (c *CloudflareImages) GetMediumURL(imageID string) string {
	return c.GetVariantURL(imageID, VariantMedium)
}

func (c *CloudflareImages) GetLargeURL(imageID string) string {
	return c.GetVariantURL(imageID, VariantLarge)
}
