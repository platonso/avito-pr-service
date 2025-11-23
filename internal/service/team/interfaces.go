package team

import (
	"context"
	"github.com/platonso/avito-pr-service/internal/domain"
)

type ServiceInterface interface {
	CreateTeam(ctx context.Context, team *domain.Team) error
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
}
