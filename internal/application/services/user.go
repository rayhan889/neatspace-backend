package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rayhan889/neatspace/internal/application/handler/dto"
	"github.com/rayhan889/neatspace/internal/domain/user/entities"
	"github.com/rayhan889/neatspace/internal/domain/user/repositories"
	"github.com/rayhan889/neatspace/pkg/apputils"
)

type UserServiceInterface interface {
	PaginationUser(c *fiber.Ctx, p *apputils.Pagination) (data []dto.UserPagination, total int, err error)
	CreateUser(ctx context.Context, user *entities.UserEntity) error
	GetUserByEmail(ctx context.Context, email string) (*entities.UserEntity, error)
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error
	IsUserExistsByID(ctx context.Context, userID uuid.UUID) bool
}

var _ UserServiceInterface = (*UserService)(nil)

type UserService struct {
	userRepo repositories.UserRepositoryInterface
}

type UserServiceOpts struct {
	UserRepo repositories.UserRepositoryInterface
}

func NewUserService(opts UserServiceOpts) *UserService {
	return &UserService{
		userRepo: opts.UserRepo,
	}
}

func (s *UserService) PaginationUser(c *fiber.Ctx, p *apputils.Pagination) (data []dto.UserPagination, total int, err error) {
	return s.userRepo.PaginationUser(c, p)
}

func (s *UserService) CreateUser(ctx context.Context, user *entities.UserEntity) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	if user.Metadata == nil {
		user.Metadata = &entities.UserMetadata{
			Timezone: "UTC",
			Role:     "user",
		}
	}

	base := strings.SplitN(user.Email, "@", 2)[0]

	re := regexp.MustCompile(`[^a-z0-9_]+`)
	sanitized := re.ReplaceAllString(strings.ToLower(base), "")
	if sanitized == "" {
		sanitized = "user"
	}
	username := sanitized

	exist, err := s.userRepo.EmailExists(ctx, user.Email)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error checking email existence: %v", err))
	}
	if exist {
		return fiber.NewError(fiber.StatusConflict, "email already exists")
	}

	user.Username = &username

	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error creating user: %v", err))
	}

	return nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*entities.UserEntity, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error getting user by email: %v", err))
	}

	return user, nil
}

func (s *UserService) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	err := s.userRepo.UpdateUserEmailVerifiedAt(ctx, userID, now)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error updating email verified: %v", err))
	}

	return nil
}

func (s *UserService) IsUserExistsByID(ctx context.Context, userID uuid.UUID) bool {
	return s.userRepo.IsUserExistsByID(ctx, userID)
}
