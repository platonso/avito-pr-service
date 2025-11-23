package user

import (
	"context"
	"github.com/platonso/avito-pr-service/internal/domain"
)

type ServiceInterface interface {
	SetUserIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetPRsByUserID(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
}
