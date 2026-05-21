package middleware

import (
	"backend/pkg/response"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ControlHubOnly ensures the request is authenticated as a ControlHub company admin.
// This middleware checks for "controlhub" role and "company_admin" level in JWT claims.
func ControlHubOnly() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get("role").(string)
			if !ok || role != "controlhub" {
				return c.JSON(http.StatusForbidden, response.Error("ControlHub access required"))
			}

			level, ok := c.Get("level").(string)
			if !ok || level != "company_admin" {
				return c.JSON(http.StatusForbidden, response.Error("Company admin access required"))
			}

			return next(c)
		}
	}
}
