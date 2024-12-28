package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	// JWT secret key'i al
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return nil, fmt.Errorf("JWT_SECRET is not set")
	}

	// Token'ı parse et
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Token'ın algoritmasını kontrol et
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// Token'ın geçerliliğini kontrol et
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Token'ın süresinin dolup dolmadığını kontrol et
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return nil, fmt.Errorf("token has expired")
			}
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
