package middleware

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(c *fiber.Ctx) error {
	tokenString := c.Cookies("AuthToken")
	if tokenString == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Authentication required"})
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("SECRET_KEY")), nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		if time.Unix(int64(claims["exp"].(float64)), 0).Before(time.Now()) {
			return c.Status(401).JSON(fiber.Map{"error": "Token expired"})
		}
		
		// Extract username and userId directly from JWT claims
		username := claims["sub"].(string)
		userID := claims["userId"].(string)
		
		c.Locals("userName", username)
		c.Locals("userID", userID)
	}

	return c.Next()
}
