package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/platonso/avito-pr-service/internal/domain"
	"github.com/platonso/avito-pr-service/internal/repository"
	"log/slog"
)

type TeamService struct {
	log      *slog.Logger
	teamRepo repository.TeamRepository
}

func NewTeamService(
	teamRepo repository.TeamRepository,
	log *slog.Logger,
) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		log:      log,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, team *domain.Team) error {
	// Validate unique users
	userIDs := make(map[string]bool)
	for _, member := range team.Members {
		if userIDs[member.ID] {
			return domain.NewError(domain.ErrCodeBadRequest, "duplicate user")
		}
		userIDs[member.ID] = true
	}

	err := s.teamRepo.CreateWithMembers(ctx, team)
	if err != nil {
		if errors.Is(err, repository.ErrTeamAlreadyExists) {
			s.log.Warn("team already exists", slog.String("team_name", team.Name))
			return domain.NewError(domain.ErrCodeTeamExists, "team_name already exists")
		}

		s.log.Error("failed to create team", slog.String("error", err.Error()))
		return fmt.Errorf("failed to create team: %w", err)
	}

	return nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := s.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, repository.ErrTeamNotFound) {
			s.log.Warn("team not found", slog.String("team_name", teamName))
			return nil, domain.NewError(domain.ErrCodeNotFound, "resource not found")
		}
		s.log.Error("failed to get team", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	return team, nil
}
