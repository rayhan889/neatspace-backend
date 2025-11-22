package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rayhan889/neatspace/internal/config"
)

func CORSMiddleware(cfg *config.Config) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins: strings.Join(cfg.App.CORSOrigins, ","),
		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodOptions,
		}, ","),
		AllowHeaders: strings.Join([]string{
			fiber.HeaderOrigin,
			fiber.HeaderContentType,
			fiber.HeaderAccept,
		}, ","),
		ExposeHeaders: strings.Join([]string{
			fiber.HeaderAccept,
			fiber.HeaderAcceptEncoding,
			fiber.HeaderAuthorization,
			fiber.HeaderCacheControl,
			fiber.HeaderConnection,
			fiber.HeaderContentLength,
			fiber.HeaderContentType,
			fiber.HeaderOrigin,
			"X-CSRF-Token",
			fiber.HeaderXRequestID,
			"Pragma",
			"User-Agent",
			"X-App-Audience",
			"X-Signature",
		}, ","),
		AllowCredentials: cfg.App.CORSCredentials,
		MaxAge:           cfg.App.CORSMaxAge,
	})
}
