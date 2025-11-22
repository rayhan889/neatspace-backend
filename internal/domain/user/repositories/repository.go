package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rayhan889/neatspace/internal/application/handler/dto"
	userEntity "github.com/rayhan889/neatspace/internal/domain/user/entities"
	"github.com/rayhan889/neatspace/pkg/apputils"
)

type UserRepositoryInterface interface {
	PaginationUser(c *fiber.Ctx, p *apputils.Pagination) (data []dto.UserPagination, total int, err error)
	CreateUser(ctx context.Context, user *userEntity.UserEntity) error
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
	GetUserByEmail(ctx context.Context, email string) (*userEntity.UserEntity, error)
}

var _ UserRepositoryInterface = (*UserRepository)(nil)

type UserRepository struct {
	pgPool *pgxpool.Pool
	logger *slog.Logger
}

func NewUserRepository(pgPool *pgxpool.Pool, logger *slog.Logger) *UserRepository {
	return &UserRepository{
		pgPool: pgPool,
		logger: logger,
	}
}

func (r *UserRepository) PaginationUser(c *fiber.Ctx, p *apputils.Pagination) (data []dto.UserPagination, total int, err error) {
	query := fmt.Sprintf(`
			SELECT display_name, username, metadata, email, email_verified_at, last_login_at 
			FROM %s
			WHERE 1=1
		`, userEntity.UserTable,
	)

	query, args := r.queryFilter(c, query)

	argPos := len(args) + 1
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, p.Limit, p.Offset)

	rows, err := r.pgPool.Query(c.Context(), query, args...)
	if err != nil {
		r.logger.Error("failed to query users", slog.String("op", "PaginationUser"), slog.String("error", err.Error()))
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var item dto.UserPagination
		var metadata userEntity.UserMetadata

		var metadataBytes []byte

		err := rows.Scan(
			&item.DisplayName,
			&item.Username,
			&metadataBytes,
			&item.Email,
			&item.EmailVerifiedAt,
			&item.LastLoginAt,
		)
		if err != nil {
			r.logger.Error("failed to scan user row", slog.String("op", "PaginationUser"), slog.String("error", err.Error()))
			return nil, 0, err
		}
		if len(metadataBytes) > 0 {
			if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
				r.logger.Error("failed to unmarshal user metadata", slog.String("op", "PaginationUser"), slog.String("error", err.Error()))
				return nil, 0, err
			}
		}

		item.Metadata = metadata

		data = append(data, item)
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM %s
		WHERE 1=1
	`, userEntity.UserTable)

	countQuery, countArgs := r.queryFilter(c, countQuery)

	err = r.pgPool.QueryRow(c.Context(), countQuery, countArgs...).Scan(&total)
	if err != nil {
		r.logger.Error("failed to count users", slog.String("op", "PaginationUser"), slog.String("error", err.Error()))
		return nil, 0, err
	}

	return data, total, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *userEntity.UserEntity) error {
	_, err := r.pgPool.Exec(ctx, fmt.Sprintf(`INSERT INTO %s (id, display_name, username, metadata, email, created_at, updated_at, email_verified_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, userEntity.UserTable),
		user.ID,
		user.DisplayName,
		user.Username,
		user.Metadata,
		user.Email,
		user.CreatedAt,
		user.UpdatedAt,
		user.EmailVerifiedAt,
	)
	if err != nil {
		r.logger.Error("failed to insert user", slog.String("op", "CreateUser"), slog.String("error", err.Error()))
		return err
	}

	r.logger.Info("user created", slog.String("op", "CreateUser"), slog.String("user_id", user.ID.String()))
	return nil
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool

	query := fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1 FROM %s WHERE LOWER(email) = LOWER($1)
		)
	`, userEntity.UserTable)

	err := r.pgPool.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		r.logger.Error("failed to check email existence", slog.String("op", "EmailExists"), slog.String("error", err.Error()))
		return false, err
	}
	return exists, nil
}

func (r *UserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool

	query := fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1 FROM %s WHERE username = $1
		)
	`, userEntity.UserTable)

	err := r.pgPool.QueryRow(ctx, query, username).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		r.logger.Error("failed to check username existence", slog.String("op", "UsernameExists"), slog.String("error", err.Error()))
		return false, err
	}
	return exists, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*userEntity.UserEntity, error) {
	var user userEntity.UserEntity
	var metadata userEntity.UserMetadata

	query := fmt.Sprintf(`
		SELECT id, display_name, username, metadata, email, email_verified_at, created_at, updated_at 
		FROM %s 
		WHERE LOWER(email) = LOWER($1)
	`, userEntity.UserTable)

	row := r.pgPool.QueryRow(ctx, query, email)
	var metadataBytes []byte

	err := row.Scan(
		&user.ID,
		&user.DisplayName,
		&user.Username,
		&metadataBytes,
		&user.Email,
		&user.EmailVerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("failed to get user by email", slog.String("op", "GetUserByEmail"), slog.String("error", err.Error()))
		return nil, err
	}

	if len(metadataBytes) > 0 {
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			r.logger.Error("failed to unmarshal user metadata", slog.String("op", "GetUserByEmail"), slog.String("error", err.Error()))
			return nil, err
		}
		user.Metadata = &metadata
	}

	return &user, nil
}

func (r *UserRepository) queryFilter(c *fiber.Ctx, baseQuery string) (string, []interface{}) {
	if baseQuery == "" {
		baseQuery = "WHERE 1=1"
	}

	var args []interface{}
	argPos := 1

	if search := c.Query("search"); search != "" {
		like := "%" + search + "%"

		fields := []string{
			fmt.Sprintf("%s.display_name", userEntity.UserTable),
			fmt.Sprintf("%s.username", userEntity.UserTable),
		}

		var conditions []string

		for _, f := range fields {
			conditions = append(conditions, fmt.Sprintf("LOWER(%s) ILIKE LOWER($%d)", f, argPos))
			args = append(args, like)
			argPos++
		}

		baseQuery += " AND (" + strings.Join(conditions, " OR ") + ")"
	}

	if role := c.Query("role"); role != "" {
		if role != "user" && role != "admin" {
			r.logger.Warn("invalid role filter value", slog.String("role", role))
			return baseQuery, args
		}

		baseQuery += fmt.Sprintf(" AND %s.metadata @> $%d", userEntity.UserTable, argPos)
		args = append(args, fmt.Sprintf(`{"role":"%s"}`, role))
		argPos++
	}

	return baseQuery, args
}
