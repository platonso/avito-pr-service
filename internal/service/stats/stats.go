package stats

import (
	"context"
	"fmt"
	"github.com/platonso/avito-pr-service/internal/domain"
	"github.com/platonso/avito-pr-service/internal/repository"
	"log/slog"
)

type Service struct {
	prRepo repository.PRRepository
	log    *slog.Logger
}

func NewService(
	prRepo repository.PRRepository,
	log *slog.Logger,
) *Service {
	return &Service{
		prRepo: prRepo,
		log:    log,
	}
}

func (s *Service) GetReviewerAssignmentsStats(ctx context.Context) ([]domain.ReviewerStat, error) {
	stats, err := s.prRepo.GetReviewerStats(ctx)
	if err != nil {
		s.log.Error(err.Error())
		return nil, fmt.Errorf("failed to get reviewer assignments stats: %w", err)
	}

	if stats == nil {
		return []domain.ReviewerStat{}, err
	}

	return stats, nil
}

func (s *Service) GetPRStats(ctx context.Context) ([]domain.PullRequestStat, error) {
	stats, err := s.prRepo.GetRRStats(ctx)
	if err != nil {
		s.log.Error(err.Error())
		return nil, fmt.Errorf("failed to get PR stats: %w", err)
	}

	if stats == nil {
		return []domain.PullRequestStat{}, err
	}

	return stats, nil
}
