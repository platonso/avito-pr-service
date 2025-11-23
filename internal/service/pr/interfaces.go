package pr

import (
	"context"
	"github.com/platonso/avito-pr-service/internal/domain"
)

type ServiceInterface interface {
	CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error)
}
