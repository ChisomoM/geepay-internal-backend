package middleware

import (
	config "backend/pkg"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWTMiddleware validates the JWT token and extracts claims into the Echo context.
// This middleware should run before CompanyMiddleware.
func JWTMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip health checks
			if c.Request().URL.Path == "/api/v1/health" {
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing authorization header"})
			}

			// Check if the header format is "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid authorization format"})
			}

			tokenString := parts[1]

			// Parse and validate the token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(cfg.JWTSecret), nil
			})

			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token"})
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token claims"})
			}

			// Check expiration
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Token expired"})
				}
			}

			// Set claims to context for downstream handlers and middleware
			c.Set("user_id", claims["sub"])
			c.Set("email", claims["email"])
			c.Set("role_slug", claims["role_slug"])
			c.Set("company_id", claims["company_id"])
			c.Set("role", claims["role"])           // ControlHub: platform-level role
			c.Set("level", claims["level"])         // ControlHub: access level
			c.Set("user_type", claims["user_type"]) // User type (e.g., controlhub_admin)

			return next(c)
		}
	}
}
