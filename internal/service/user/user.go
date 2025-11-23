package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/platonso/avito-pr-service/internal/domain"
	"github.com/platonso/avito-pr-service/internal/repository"
	"log/slog"
)

type Service struct {
	userRepo repository.UserRepository
	log      *slog.Logger
}

func NewService(
	userRepo repository.UserRepository,
	log *slog.Logger,
) *Service {
	return &Service{
		userRepo: userRepo,
		log:      log,
	}
}

func (s *Service) SetUserIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	err := s.userRepo.SetIsActive(ctx, userID, isActive)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.log.Warn("user not found", slog.String("user_id", userID))
			return nil, domain.NewError(domain.ErrCodeNotFound, "resource not found")
		}
		s.log.Error("failed to update user status", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("failed to get updated user", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get updated user: %w", err)
	}

	return user, nil
}

func (s *Service) GetPRsByUserID(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	// Check user existence
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.log.Warn("user not found", slog.String("user_id", userID))
			return nil, domain.NewError(domain.ErrCodeNotFound, "resource not found")
		}
		s.log.Error("failed to get user", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	prs, err := s.userRepo.GetPRsByUserID(ctx, userID)
	if err != nil {
		s.log.Error("failed to get user PRs", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get user PRs: %w", err)
	}
	return prs, nil
}
