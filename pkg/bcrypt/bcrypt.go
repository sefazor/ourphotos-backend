package bcrypt

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	DefaultCost = 10
)

// HashPassword şifreyi hashler
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}
	return string(hashedBytes), nil
}

// ComparePassword hashlenen şifre ile plain text şifreyi karşılaştırır
func ComparePassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("password comparison failed: %v", err)
	}
	return nil
}

// VerifyHash verilen hash'in geçerli bir bcrypt hash'i olup olmadığını kontrol eder
func VerifyHash(hash string) bool {
	return len(hash) == 60 && hash[0:2] == "$2"
}

// Debug fonksiyonu - sadece geliştirme aşamasında kullanılmalı
func DebugHashAndCompare(password string) {
	fmt.Printf("\nDebug Hash and Compare:\n")
	fmt.Printf("Original password: %s\n", password)

	hash, err := HashPassword(password)
	if err != nil {
		fmt.Printf("Hashing failed: %v\n", err)
		return
	}
	fmt.Printf("Generated hash: %s\n", hash)

	err = ComparePassword(hash, password)
	if err != nil {
		fmt.Printf("Comparison failed: %v\n", err)
		return
	}
	fmt.Printf("Password verification successful\n")
}
