package middlewares

import "github.com/gofiber/fiber/v2"

func SecurityHeadersMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// X-Content-Type-Options - prevents MIME type sniffing
		c.Response().Header.Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options - prevents clickjacking
		c.Response().Header.Set("X-Frame-Options", "DENY")

		// Strict-Transport-Security (HSTS) - forces HTTPS
		c.Response().Header.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content-Security-Policy - allow scripts from self and trusted CDN
		c.Response().Header.Set("Content-Security-Policy",
			"default-src 'none'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net; "+
				"script-src-elem 'self' 'unsafe-inline' https://cdn.jsdelivr.net; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data:; "+
				"font-src 'self' https://fonts.scalar.com; "+
				"connect-src 'self' https:; "+
				"frame-ancestors 'none'")

		return c.Next()
	}
}
