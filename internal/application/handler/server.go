package handler

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/alexliesenfeld/health"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rayhan889/neatspace/docs"
	"github.com/rayhan889/neatspace/internal/config"
	"github.com/rayhan889/neatspace/web"

	scalar "github.com/bdpiprava/scalar-go"
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

	// Register specific routes BEFORE catch-all routes
	// health check route (allow direct hit instead of redirect to SPA)
	fiberApp.Get("/healthz", h.HealthCheckHandler)
	// API docs + OpenAPI spec
	fiberApp.Get("/api-docs", h.APIDocsHandler)
	fiberApp.Get("/api/openapi.json", h.OpenAPISpecHandler)

	// Serve index.html for root and all non-static paths (catch-all LAST)
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

// @Summary		    Service healthcheck
// @Description	    Checks the health of the service
// @Tags	        General Information
// @Router		    /healthz [get]
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

func (h *ServerHandler) OpenAPISpecHandler(c *fiber.Ctx) error {
	cfg := config.Get()

	if !cfg.IsAPIDocsEnabled() {
		return fiber.NewError(fiber.StatusUnauthorized, "API Docs are disabled")
	}

	specByte, err := docs.SwaggerFS.ReadFile("swagger.json")
	if err != nil {
		h.Logger.Error("failed to read swagger.json file", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError)
	}

	return c.Send(specByte)
}

func (h *ServerHandler) APIDocsHandler(c *fiber.Ctx) error {
	cfg := config.Get()

	if !cfg.IsAPIDocsEnabled() {
		return fiber.NewError(fiber.StatusUnauthorized, "API Docs are disabled")
	}

	scheme := "http"
	if c.Context().IsTLS() {
		scheme = "https"
	}
	specURL := fmt.Sprintf("%s://%s/api/openapi.json", scheme, c.Hostname())

	// Generate HTML content using scalargo NewV2 with spec URL
	htmlContent, err := scalar.NewV2(
		scalar.WithSpecURL(specURL),
		scalar.WithAuthenticationOpts(
			scalar.WithCustomSecurity(),
			scalar.WithPreferredSecurityScheme("bearerAuth"),
			scalar.WithHTTPBearerToken("your-bearer-token-here"),
		),
		scalar.WithServers(scalar.Server{
			URL:         fmt.Sprintf("%s://%s", scheme, c.Hostname()),
			Description: "Default server",
		}),
		scalar.WithHiddenClients(
			"libcurl",       // C
			"httpclient",    // CSharp
			"restsharp",     // CSharp
			"clj_http",      // Clojure
			"http",          // Dart
			"native",        // Go & Ruby
			"http1.1",       // HTTP
			"asynchttp",     // Java
			"nethttp",       // Java
			"okhttp",        // Java & Kotlin
			"unirest",       // Java
			"jquery",        // JavaScript
			"xhr",           // JavaScript
			"nsurlsession",  // Objective-C & Swift
			"cohttp",        // OCaml
			"guzzle",        // PHP
			"webrequest",    // Powershell
			"restmethod",    // Powershell
			"http.client",   // Python
			"requests",      // Python
			"python3",       // Python
			"HTTPX (Async)", // Python
			"httr",          // R
			"request",       // Unknown
			"http1",         // Unknown
			"http2",         // Unknown
		),
		scalar.WithLayout(scalar.LayoutModern),
		scalar.WithHideDarkModeToggle(),
		scalar.WithHideModels(),
		scalar.WithDarkMode(),
		scalar.WithTheme(scalar.ThemeMoon),
		scalar.WithOverrideCSS(`
            aside div.flex.items-center:has(a[href*="scalar.com"]) { display: none !important; }
            .scalar-app { font-family: -apple-system, BlinkMacSystemFont, Aptos, "Segoe UI", Roboto, sans-serif; }
        `),
	)

	if err != nil {
		h.Logger.Error("failed to generate Scalar API docs", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("failed to generate API docs: %v", err))
	}

	c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return c.SendString(htmlContent)
}

// getFileSystem always uses embed.FS for static assets.
func getFileSystem(embedFS embed.FS, subdir string) http.FileSystem {
	fsys, err := fs.Sub(embedFS, subdir)
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}
