package notification

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/mail"
	"strings"
	"text/template"

	gomail "gopkg.in/mail.v2"
)

type Mailer struct {
	host       string
	port       int
	username   string
	password   string
	fromName   string
	fromAddr   string
	templateFS fs.FS

	logger *slog.Logger
}

type MailerOpts struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string

	FromName string
	FromAddr string

	TemplateFS fs.FS

	Logger *slog.Logger
}

// sanitizeHeader trims whitespace/quotes and strips CR/LF to prevent header injection.
func sanitizeHeader(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}

func NewMailer(opts MailerOpts) (*Mailer, error) {
	fromAddr := sanitizeHeader(opts.FromAddr)
	fromName := sanitizeHeader(opts.FromName)

	m := &Mailer{
		host:       opts.SMTPHost,
		port:       opts.SMTPPort,
		username:   opts.SMTPUsername,
		password:   opts.SMTPPassword,
		fromName:   fromName,
		fromAddr:   fromAddr,
		templateFS: opts.TemplateFS,
		logger:     opts.Logger,
	}

	if opts.SMTPPort != 0 {
		m.port = opts.SMTPPort
	}

	// validate required fields
	if m.host == "" {
		return nil, errors.New("smtp host is required")
	}
	if m.fromAddr == "" {
		return nil, errors.New("from address is required")
	}

	// validate email format
	if _, err := mail.ParseAddress(m.fromAddr); err != nil {
		return nil, errors.New("invalid from address")
	}

	// ensure fromName doesn't contain angle brackets (avoid header injection)
	m.fromName = strings.ReplaceAll(m.fromName, "<", "")
	m.fromName = strings.ReplaceAll(m.fromName, ">", "")

	// check templateFS was provided
	if m.templateFS == nil {
		return nil, errors.New("template FS is required")
	}

	// ensure logger
	if m.logger == nil {
		m.logger = slog.New(slog.NewJSONHandler(nil, &slog.HandlerOptions{}))
	}

	m.logger.Debug("mailer configured", "host", m.host, "port", m.port, "from", m.fromAddr)
	return m, nil
}

func (m *Mailer) SendMail(ctx context.Context, to []string, subject, templateName string, data any) error {
	if len(to) == 0 {
		return errors.New("no recipient addresses provided")
	}

	body, err := m.formatEmailHTML(templateName, data)
	if err != nil {
		return err
	}

	msg := gomail.NewMessage()

	msg.SetHeader("From", fmt.Sprintf("%s <%s>", m.fromName, m.fromAddr))
	msg.SetHeader("To", to...)
	msg.SetHeader("Subject", subject)

	msg.SetBody("text/plain", body)

	dialer := gomail.NewDialer(m.host, m.port, m.username, m.password)

	if err := dialer.DialAndSend(msg); err != nil {
		m.logger.Error("failed to send email", "err", err, "to", to, "subject", subject)
		return err
	}

	m.logger.Debug("email sent", "to", to, "subject", subject)
	return nil
}

// formatEmailHTML loads the named template from the embedded FS and executes it with data.
// Template files should be located under "emails/" in the embedded FS (see templates/embed.go).
func (m *Mailer) formatEmailHTML(templateName string, data any) (string, error) {
	if templateName == "" {
		return "", errors.New("template name required")
	}

	// Load the specific template file from the FS (e.g. "emails/welcome.html").
	tplPath := "emails/" + templateName
	t, err := template.ParseFS(m.templateFS, tplPath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
