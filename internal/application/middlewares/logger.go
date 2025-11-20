package middlewares

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rayhan889/intern-payroll/pkg/apputils"
)

// LoggerMiddleware returns an Fiber middleware that logs HTTP requests using slog.
func LoggerMiddleware(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate latency
		stop := time.Now()
		latency := stop.Sub(start)

		// Get request details
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		clientIP := c.IP()
		realUserAgent := string(c.Request().Header.UserAgent())
		userAgent := apputils.SummarizeUserAgent(realUserAgent)

		// Get or generate request ID
		requestID := c.Get(fiber.HeaderXRequestID)
		if requestID == "" {
			requestID = c.GetRespHeader(fiber.HeaderXRequestID)
		}

		// Get content length
		bytesIn := len(c.Request().Body())
		bytesOut := len(c.Response().Body())

		logAttrs := []slog.Attr{
			slog.Int("status", status),
			slog.String("method", method),
			slog.String("request_id", requestID),
			slog.String("path", path),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", userAgent),
			slog.String("duration", formatLatency(latency)),
			slog.Int("bytes_in", bytesIn),
			slog.Int("bytes_out", bytesOut),
		}

		// Add query parameters if present
		if c.Request().URI().QueryString() != nil && len(c.Request().URI().QueryString()) > 0 {
			logAttrs = append(logAttrs, slog.String("query", string(c.Request().URI().QueryString())))
		}

		// Log with appropriate level based on status code
		switch {
		case status >= 500:
			logger.Error("HTTP Request", attrsToArgs(logAttrs)...)
		case status >= 400:
			logger.Warn("HTTP Request", attrsToArgs(logAttrs)...)
		default:
			logger.Info("HTTP Request", attrsToArgs(logAttrs)...)
		}

		return err
	}
}

func formatLatency(d time.Duration) string {
	ns := d.Nanoseconds()
	switch {
	case ns < 1_000:
		return fmt.Sprintf("%dns", ns)
	case ns < 1_000_000:
		return fmt.Sprintf("%.2fµs", float64(ns)/1_000)
	case ns < 1_000_000_000:
		return fmt.Sprintf("%.2fms", float64(ns)/1_000_000)
	default:
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}

func attrsToArgs(attrs []slog.Attr) []any {
	args := make([]any, 0, len(attrs)*2)
	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value.Any())
	}
	return args
}
