package repository

import (
	"context"
	"github.com/platonso/avito-pr-service/internal/domain"
	"time"
)

type TeamRepository interface {
	CreateWithMembers(ctx context.Context, team *domain.Team) error
	GetByName(ctx context.Context, teamName string) (*domain.Team, error)
	GetByUserID(ctx context.Context, userID string) (*domain.Team, error)
	Exists(ctx context.Context, teamName string) (bool, error)
}

type UserRepository interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) error
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	GetPRsByUserID(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
}

type PRRepository interface {
	Create(ctx context.Context, pullRequest *domain.PullRequest) error
	Merge(ctx context.Context, prID string, mergedAt time.Time) error
	GetByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	GetReviewersIDs(ctx context.Context, prID string) ([]string, error)
	ChangeReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
	Exists(ctx context.Context, prID string) (bool, error)
}
