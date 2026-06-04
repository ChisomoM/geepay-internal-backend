package middleware

import (
	"backend/db"
	config "backend/pkg"
	"backend/pkg/response"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// MerchantJWTMiddleware validates merchant portal JWTs and injects the global DB.
// Merchant routes scope queries by merchant_id rather than company_id.
func MerchantJWTMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, response.Error("missing authorization header"))
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, response.Error("invalid authorization format"))
			}

			token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(cfg.JWTSecret), nil
			})
			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, response.Error("invalid or expired token"))
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, response.Error("invalid token claims"))
			}

			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					return c.JSON(http.StatusUnauthorized, response.Error("token expired"))
				}
			}

			userType, _ := claims["user_type"].(string)
			if userType != "merchant" {
				return c.JSON(http.StatusForbidden, response.Error("merchant portal access only"))
			}

			c.Set("merchant_id", claims["merchant_id"])
			c.Set("merchant_email", claims["email"])
			// Inject global DB — service layer scopes by merchant_id
			c.Set("db", db.GetDB())

			return next(c)
		}
	}
}
