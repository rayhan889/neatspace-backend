package handler

import (
	"context"
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/alexliesenfeld/health"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rayhan889/neatspace/web"
)

// ServerHandler holds dependencies for HTTP handlers.
type ServerHandler struct {
	PGPool *pgxpool.Pool
	Logger *slog.Logger
	WebFS  embed.FS
}

// NewServerHandler creates a new ServerHandler.
func NewServerHandler(pgPool *pgxpool.Pool, logger *slog.Logger) *ServerHandler {
	return &ServerHandler{
		PGPool: pgPool,
		Logger: logger,
		WebFS:  web.WebDir,
	}
}

func (h *ServerHandler) RegisterRoutes(fiberApp *fiber.App) {
	staticFS := getFileSystem(h.WebFS, "static")

	// serve static files under /static/*
	// fiber.Static("/static", ".", fiber.Static{

	// })

	// health check route (allow direct hit instead of redirect to SPA)
	fiberApp.Get("/healthz", h.HealthCheckHandler)

	// API docs + OpenAPI spec
	// fiberApp.Get("/api-docs", h.APIDocsHandler)
	// fiberApp.Get("/api/openapi.json", h.OpenAPISpecHandler)

	// Serve index.html for root and all non-static paths
	fiberApp.Get("/", h.RootHandler(staticFS))
	fiberApp.All("/*", h.RootHandler(staticFS))
}

func (h *ServerHandler) RootHandler(staticFS http.FileSystem) fiber.Handler {
	return func(c *fiber.Ctx) error {
		upath := c.Path()

		// Ignore static/healthz routes
		if strings.HasPrefix(upath, "/static/") || upath == "/healthz" {
			return fiber.ErrNotFound
		}

		f, err := staticFS.Open("index.html")
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "index.html not found")
		}
		defer func() {
			if cerr := f.Close(); cerr != nil {
				h.Logger.Warn("failed to close index.html file", "err", cerr)
			}
		}()

		// Get file info (for modtime)
		stat, err := f.Stat()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to stat index.html")
		}

		c.Set(fiber.HeaderContentType, "text/html; charset=utf-8")

		// Serve file through io.Reader so Go can stream it
		c.Response().Header.Set(fiber.HeaderContentType, "text/html; charset=utf-8")

		return c.SendStream(f, int(stat.Size()))
	}
}

func (h *ServerHandler) HealthCheckHandler(c *fiber.Ctx) error {
	hc := health.NewChecker(
		health.WithCacheDuration(10*time.Second),
		health.WithTimeout(5*time.Second),
		health.WithCheck(health.Check{
			Name:    "database",
			Timeout: 2 * time.Second,
			Check: func(ctx context.Context) error {
				return h.PGPool.Ping(ctx)
			},
		}),
	)

	handler := health.NewHandler(hc)

	return adaptor.HTTPHandler(handler)(c)
}

// getFileSystem always uses embed.FS for static assets.
func getFileSystem(embedFS embed.FS, subdir string) http.FileSystem {
	fsys, err := fs.Sub(embedFS, subdir)
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}
