package config

// JWTAlgorithm is a typesafe enum for JWT algorithm
// Supported values: "HS256", "RS256"
type JWTAlgorithm string

const (
	JWTAlgorithmHS256 JWTAlgorithm = "HS256"
	JWTAlgorithmRS256 JWTAlgorithm = "RS256"
)

type Config struct {
	App      AppConfig      `env:",squash"`
	Database DatabaseConfig `env:",squash"`
	Logging  LoggingConfig  `env:",squash"`
	Mailer   MailerConfig   `env:",squash"`
}

type AppConfig struct {
	Mode               string       `env:"APP_MODE"` // development|production
	BaseURL            string       `env:"APP_BASE_URL"`
	JWTSecretKey       string       `env:"JWT_SECRET_KEY"`
	JWTAlgorithm       JWTAlgorithm `env:"JWT_ALGORITHM"`
	ServerHost         string       `env:"SERVER_HOST"`
	ServerPort         int          `env:"SERVER_PORT"`
	CORSOrigins        []string     `env:"CORS_ORIGINS"`
	CORSMaxAge         int          `env:"CORS_MAX_AGE"`
	CORSCredentials    bool         `env:"CORS_CREDENTIALS"`
	RateLimitEnabled   bool         `env:"RATE_LIMIT_ENABLED"`
	RateLimitRequests  int          `env:"RATE_LIMIT_REQUESTS"`
	RateLimitBurstSize int          `env:"RATE_LIMIT_BURST_SIZE"`
	EnableAPIDocs      bool         `env:"ENABLE_API_DOCS"`
}

type DatabaseConfig struct {
	PostgresURL string `env:"DATABASE_URL"`
	MaxPoolSize int    `env:"PG_MAX_POOL_SIZE"`
	MaxRetries  int    `env:"PG_MAX_RETRIES"`
}

type LoggingConfig struct {
	Level   string `env:"LOG_LEVEL"`
	Format  string `env:"LOG_FORMAT"`
	NoColor bool   `env:"LOG_NO_COLOR"`
}

type MailerConfig struct {
	SMTPHost     string `env:"SMTP_HOST"`
	SMTPPort     int    `env:"SMTP_PORT"`
	SMTPUsername string `env:"SMTP_USERNAME"`
	SMTPPassword string `env:"SMTP_PASSWORD"`
	SenderName   string `env:"SMTP_SENDER_NAME"`
	SenderEmail  string `env:"SMTP_SENDER_EMAIL"`
	SMTPSecure   bool   `env:"SMTP_SECURE"`
}
