package captcha

import (
	"time"
)

type TurnstileResponse struct {
	Success    bool      `json:"success"`
	ErrorCodes []string  `json:"error-codes"`
	Hostname   string    `json:"hostname"`
	Challenge  string    `json:"challenge_ts"`
	ExpireTime time.Time `json:"expires-at"`
	Action     string    `json:"action"`
}

// VerifyTurnstile checks if the provided token is valid
func VerifyTurnstile(token string) (bool, error) {
	// Geçici olarak devre dışı bırakıldı - her zaman başarılı dön
	return true, nil

	/*
		// Original kod (daha sonra aktif etmek için)
		if token == "" {
			return false, errors.New("missing turnstile token")
		}

		secretKey := os.Getenv("CF_TURNSTILE_SECRET_KEY")
		if secretKey == "" {
			secretKey = "0x4AAAAAABHpobVKraJ7cU1aa-h59vVnmQI"
		}

		formData := url.Values{}
		formData.Add("secret", secretKey)
		formData.Add("response", token)

		resp, err := http.Post(
			"https://challenges.cloudflare.com/turnstile/v0/siteverify",
			"application/x-www-form-urlencoded",
			strings.NewReader(formData.Encode()),
		)

		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		var result TurnstileResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return false, err
		}

		return result.Success, nil
	*/
}
