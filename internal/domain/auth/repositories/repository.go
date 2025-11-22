package repositories

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	authEntity "github.com/rayhan889/neatspace/internal/domain/auth/entities"
)

type AuthRepositoryInterface interface {
	FindAllOneTimeTokens(ctx context.Context) ([]authEntity.OneTimeToken, error)
	CreateOneTimeToken(ctx context.Context, token *authEntity.OneTimeToken) error
	UpdateOneTImeTokenLastSentAt(ctx context.Context, tokenID uuid.UUID, lastSentAt time.Time) error
	DeleteOneTimeToken(ctx context.Context, tokenID uuid.UUID) error
}

var _ AuthRepositoryInterface = (*AuthRepository)(nil)

type AuthRepository struct {
	pgPool *pgxpool.Pool
	logger *slog.Logger
}

func NewAuthRepository(pgPool *pgxpool.Pool, logger *slog.Logger) *AuthRepository {
	return &AuthRepository{
		pgPool: pgPool,
		logger: logger,
	}
}

func (r *AuthRepository) FindAllOneTimeTokens(ctx context.Context) ([]authEntity.OneTimeToken, error) {
	query := fmt.Sprintf(`SELECT id, user_id, subject, token_hash, relates_to, metadata, created_at, expires_at, last_sent_at 
		FROM %s`, authEntity.OneTimeTokenTable,
	)

	rows, err := r.pgPool.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to query one-time tokens", slog.String("op", "FindAllOneTimeTokens"), slog.String("error", err.Error()))
		return nil, err
	}

	defer rows.Close()

	var tokens []authEntity.OneTimeToken
	for rows.Next() {
		var token authEntity.OneTimeToken
		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.Subject,
			&token.TokenHash,
			&token.RelatesTo,
			&token.Metadata,
			&token.CreatedAt,
			&token.ExpiresAt,
			&token.LastSentAt,
		)
		if err != nil {
			r.logger.Error("failed to scan one-time token", slog.String("op", "FindAllOneTimeTokens"), slog.String("error", err.Error()))
			return nil, err
		}

		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (r *AuthRepository) CreateOneTimeToken(ctx context.Context, token *authEntity.OneTimeToken) error {
	_, err := r.pgPool.Exec(ctx, fmt.Sprintf(`INSERT INTO %s (id, user_id, subject, token_hash, relates_to, metadata, created_at, expires_at, last_sent_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`, authEntity.OneTimeTokenTable),
		token.ID,
		token.UserID,
		token.Subject,
		token.TokenHash,
		token.RelatesTo,
		token.Metadata,
		token.CreatedAt,
		token.ExpiresAt,
		token.LastSentAt,
	)
	if err != nil {
		r.logger.Error("failed to insert one-time token", slog.String("op", "CreateOneTimeToken"), slog.String("error", err.Error()))
		return err
	}

	r.logger.Info("one-time token created", slog.String("op", "CreateOneTimeToken"), slog.String("token_id", token.ID.String()))
	return nil
}

func (r *AuthRepository) UpdateOneTImeTokenLastSentAt(ctx context.Context, tokenID uuid.UUID, lastSentAt time.Time) error {
	query := fmt.Sprintf(`UPDATE %s SET last_sent_at = $1 WHERE id = $2`, authEntity.OneTimeTokenTable)

	cmd, err := r.pgPool.Exec(ctx, query, lastSentAt, tokenID)
	if err != nil {
		r.logger.Error("failed to update one-time token last_sent_at", slog.String("op", "UpdateOneTImeTokenLastSentAt"), slog.String("error", err.Error()))
		return err
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no one-time token found to update", slog.String("op", "UpdateOneTImeTokenLastSentAt"), slog.String("token_id", tokenID.String()))
		return fmt.Errorf("no one-time token found with id: %s", tokenID.String())
	}

	r.logger.Info("one-time token last_sent_at updated", slog.String("op", "UpdateOneTImeTokenLastSentAt"), slog.String("token_id", tokenID.String()))
	return nil
}

func (r *AuthRepository) DeleteOneTimeToken(ctx context.Context, tokenID uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, authEntity.OneTimeTokenTable)

	cmd, err := r.pgPool.Exec(ctx, query, tokenID)
	if err != nil {
		r.logger.Error("failed to delete one-time token", slog.String("op", "DeleteOneTimeToken"), slog.String("error", err.Error()))
		return err
	}
	if cmd.RowsAffected() == 0 {
		r.logger.Warn("no one-time token found to delete", slog.String("op", "DeleteOneTimeToken"), slog.String("token_id", tokenID.String()))
		return fmt.Errorf("no one-time token found with id: %s", tokenID.String())
	}

	r.logger.Info("one-time token deleted", slog.String("op", "DeleteOneTimeToken"), slog.String("token_id", tokenID.String()))
	return nil
}
