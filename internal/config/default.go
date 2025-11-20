package config

func DefaultConfig() Config {
	return Config{
		App: AppConfig{
			Mode:               "development",
			BaseURL:            "http://localhost:8000",
			JWTSecretKey:       "_THIS_IS_DEFAULT_JWT_SECRET_KEY_",
			JWTAlgorithm:       JWTAlgorithmHS256,
			ServerHost:         "0.0.0.0",
			ServerPort:         8000,
			CORSOrigins:        []string{"*"},
			CORSMaxAge:         300,
			CORSCredentials:    true,
			RateLimitEnabled:   true,
			RateLimitRequests:  20,
			RateLimitBurstSize: 60,
			EnableAPIDocs:      true,
		},
		Database: DatabaseConfig{
			PostgresURL: "postgresql://postgres:securedb@localhost:5432/postgres?sslmode=disable",
			MaxPoolSize: 10,
			MaxRetries:  5,
		},
		Logging: LoggingConfig{
			Level:   "info",
			Format:  "pretty",
			NoColor: false,
		},
		Mailer: MailerConfig{
			SMTPHost:     "localhost",
			SMTPPort:     1025,
			SMTPUsername: "",
			SMTPPassword: "",
			SenderName:   "\"System Mailer\"",
			SenderEmail:  "\"mailer@example.com\"",
			SMTPSecure:   false,
		},
	}
}
