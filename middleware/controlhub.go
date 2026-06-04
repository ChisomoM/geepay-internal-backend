package middleware

// ControlHubOnly has been removed. Platform-level actions are now gated by
// the super_admin role via the standard RBAC middleware.

// func ControlHubOnly() echo.MiddlewareFunc {
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {
// 			role, ok := c.Get("role").(string)
// 			if !ok || role != "controlhub" {
// 				return c.JSON(http.StatusForbidden, response.Error("ControlHub access required"))
// 			}
// 			level, ok := c.Get("level").(string)
// 			if !ok || level != "company_admin" {
// 				return c.JSON(http.StatusForbidden, response.Error("Company admin access required"))
// 			}
// 			return next(c)
// 		}
// 	}
// }
