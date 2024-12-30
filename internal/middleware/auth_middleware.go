package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	jwtPkg "github.com/sefazor/ourphotos-backend/pkg/jwt"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")
		fmt.Printf("\nAuth Debug:\n")
		fmt.Printf("Auth Header: %s\n", authHeader)
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authorization header is required",
			})
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid authorization header format",
			})
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate the token
		claims, err := jwtPkg.ValidateToken(tokenString)
		if err != nil {
			fmt.Printf("Token validation failed: %v\n", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid token",
			})
		}

		fmt.Printf("Token Claims: %+v\n", claims)

		// Güvenli bir şekilde userID ve email'i al
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			fmt.Printf("Invalid user_id in claims\n")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid user ID in token",
			})
		}
		userID := uint(userIDFloat)

		userEmail, ok := claims["email"].(string)
		if !ok {
			fmt.Printf("Invalid email in claims\n")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid email in token",
			})
		}

		fmt.Printf("User ID: %d, Email: %s\n", userID, userEmail)

		c.Locals("userID", userID)
		c.Locals("userEmail", userEmail)

		return c.Next()
	}
}
