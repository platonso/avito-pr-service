package team

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/platonso/avito-pr-service/internal/domain"
	"github.com/platonso/avito-pr-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockTeamRepository struct {
	CreateWithMembersFunc func(ctx context.Context, team *domain.Team) error
	GetByNameFunc         func(ctx context.Context, teamName string) (*domain.Team, error)
	GetByUserIDFunc       func(ctx context.Context, userID string) (*domain.Team, error)
	ExistsFunc            func(ctx context.Context, teamName string) (bool, error)
}

func (m *MockTeamRepository) CreateWithMembers(ctx context.Context, team *domain.Team) error {
	if m.CreateWithMembersFunc != nil {
		return m.CreateWithMembersFunc(ctx, team)
	}
	return nil
}

func (m *MockTeamRepository) GetByName(ctx context.Context, teamName string) (*domain.Team, error) {
	if m.GetByNameFunc != nil {
		return m.GetByNameFunc(ctx, teamName)
	}
	return nil, nil
}

func (m *MockTeamRepository) GetByUserID(ctx context.Context, userID string) (*domain.Team, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockTeamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, teamName)
	}
	return false, nil
}

func getTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestService_CreateTeam(t *testing.T) {
	tests := []struct {
		name          string
		team          *domain.Team
		setupMocks    func(*MockTeamRepository)
		expectedError *domain.Error
	}{
		{
			name: "successful creation",
			team: &domain.Team{
				Name: "team-1",
				Members: []domain.TeamMember{
					{ID: "user-1", Name: "User 1"},
					{ID: "user-2", Name: "User 2"},
				},
			},
			setupMocks: func(teamRepo *MockTeamRepository) {
				teamRepo.CreateWithMembersFunc = func(ctx context.Context, team *domain.Team) error {
					return nil
				}
			},
		},
		{
			name: "duplicate users",
			team: &domain.Team{
				Name: "team-1",
				Members: []domain.TeamMember{
					{ID: "user-1", Name: "User 1"},
					{ID: "user-1", Name: "User 1"},
				},
			},
			setupMocks:    func(teamRepo *MockTeamRepository) {},
			expectedError: domain.NewError(domain.ErrCodeBadRequest, "duplicate user"),
		},
		{
			name: "team already exists",
			team: &domain.Team{
				Name: "team-1",
				Members: []domain.TeamMember{
					{ID: "user-1", Name: "User 1"},
				},
			},
			setupMocks: func(teamRepo *MockTeamRepository) {
				teamRepo.CreateWithMembersFunc = func(ctx context.Context, team *domain.Team) error {
					return repository.ErrTeamAlreadyExists
				}
			},
			expectedError: domain.NewError(domain.ErrCodeTeamExists, "team_name already exists"),
		},
		{
			name: "repository error",
			team: &domain.Team{
				Name: "team-1",
				Members: []domain.TeamMember{
					{ID: "user-1", Name: "User 1"},
				},
			},
			setupMocks: func(teamRepo *MockTeamRepository) {
				teamRepo.CreateWithMembersFunc = func(ctx context.Context, team *domain.Team) error {
					return errors.New("database connection error")
				}
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teamRepo := &MockTeamRepository{}
			tt.setupMocks(teamRepo)

			service := NewService(teamRepo, getTestLogger())
			err := service.CreateTeam(context.Background(), tt.team)
			switch {
			case tt.expectedError != nil:
				require.Error(t, err)
				var domainErr *domain.Error
				require.True(t, errors.As(err, &domainErr))
				assert.Equal(t, tt.expectedError.Code, domainErr.Code)
				assert.Equal(t, tt.expectedError.Message, domainErr.Message)
			case tt.name == "repository error":
				require.Error(t, err)
				assert.Contains(t, err.Error(), "failed to create team")
			default:
				require.NoError(t, err)
			}
		})
	}
}

func TestService_GetTeam(t *testing.T) {
	tests := []struct {
		name           string
		teamName       string
		setupMocks     func(*MockTeamRepository)
		expectedError  *domain.Error
		validateResult func(*testing.T, *domain.Team)
	}{
		{
			name:     "successful get",
			teamName: "team-1",
			setupMocks: func(teamRepo *MockTeamRepository) {
				active := true
				teamRepo.GetByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						Name: teamName,
						Members: []domain.TeamMember{
							{ID: "user-1", Name: "User 1", IsActive: &active},
							{ID: "user-2", Name: "User 2", IsActive: &active},
						},
					}, nil
				}
			},
			validateResult: func(t *testing.T, team *domain.Team) {
				assert.NotNil(t, team)
				assert.Equal(t, "team-1", team.Name)
				assert.Len(t, team.Members, 2)
			},
		},
		{
			name:     "team not found",
			teamName: "team-1",
			setupMocks: func(teamRepo *MockTeamRepository) {
				teamRepo.GetByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return nil, repository.ErrTeamNotFound
				}
			},
			expectedError: domain.NewError(domain.ErrCodeNotFound, "resource not found"),
		},
		{
			name:     "repository error",
			teamName: "team-1",
			setupMocks: func(teamRepo *MockTeamRepository) {
				teamRepo.GetByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return nil, errors.New("database connection error")
				}
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teamRepo := &MockTeamRepository{}
			tt.setupMocks(teamRepo)

			service := NewService(teamRepo, getTestLogger())
			result, err := service.GetTeam(context.Background(), tt.teamName)

			switch {
			case tt.expectedError != nil:
				require.Error(t, err)
				var domainErr *domain.Error
				require.True(t, errors.As(err, &domainErr))
				assert.Equal(t, tt.expectedError.Code, domainErr.Code)
				assert.Nil(t, result)

			case tt.name == "repository error":
				require.Error(t, err)
				assert.Contains(t, err.Error(), "failed to get team")
				assert.Nil(t, result)

			default:
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}
