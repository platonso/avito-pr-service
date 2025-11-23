package pr

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/platonso/avito-pr-service/internal/domain"
	"github.com/platonso/avito-pr-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockPRRepository struct {
	CreateFunc         func(ctx context.Context, pr *domain.PullRequest) error
	MergeFunc          func(ctx context.Context, prID string, mergedAt time.Time) error
	GetByIDFunc        func(ctx context.Context, prID string) (*domain.PullRequest, error)
	ChangeReviewerFunc func(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
}

func (m *MockPRRepository) Create(ctx context.Context, pr *domain.PullRequest) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, pr)
	}
	return nil
}

func (m *MockPRRepository) Merge(ctx context.Context, prID string, mergedAt time.Time) error {
	if m.MergeFunc != nil {
		return m.MergeFunc(ctx, prID, mergedAt)
	}
	return nil
}

func (m *MockPRRepository) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, prID)
	}
	return nil, nil
}

func (m *MockPRRepository) ChangeReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	if m.ChangeReviewerFunc != nil {
		return m.ChangeReviewerFunc(ctx, prID, oldReviewerID, newReviewerID)
	}
	return nil
}

func (m *MockPRRepository) GetReviewersIDs(ctx context.Context, prID string) ([]string, error) {
	return nil, nil
}
func (m *MockPRRepository) Exists(ctx context.Context, prID string) (bool, error) { return false, nil }
func (m *MockPRRepository) GetReviewerStats(ctx context.Context) ([]domain.ReviewerStat, error) {
	return nil, nil
}
func (m *MockPRRepository) GetRRStats(ctx context.Context) ([]domain.PullRequestStat, error) {
	return nil, nil
}

type MockTeamRepository struct {
	GetByUserIDFunc func(ctx context.Context, userID string) (*domain.Team, error)
	GetByNameFunc   func(ctx context.Context, teamName string) (*domain.Team, error)
}

func (m *MockTeamRepository) GetByUserID(ctx context.Context, userID string) (*domain.Team, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockTeamRepository) GetByName(ctx context.Context, teamName string) (*domain.Team, error) {
	if m.GetByNameFunc != nil {
		return m.GetByNameFunc(ctx, teamName)
	}
	return nil, nil
}

func (m *MockTeamRepository) CreateWithMembers(ctx context.Context, team *domain.Team) error {
	return nil
}
func (m *MockTeamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	return false, nil
}

type MockUserRepository struct {
	GetByIDFunc func(ctx context.Context, userID string) (*domain.User, error)
}

func (m *MockUserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	return nil
}
func (m *MockUserRepository) GetPRsByUserID(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	return nil, nil
}

func getTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestService_CreatePullRequest(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(prRepo *MockPRRepository, teamRepo *MockTeamRepository, userRepo *MockUserRepository)
		expectedError *domain.Error
	}{
		{
			name: "successful creation",
			setupMocks: func(prRepo *MockPRRepository, teamRepo *MockTeamRepository, userRepo *MockUserRepository) {
				active := true
				userRepo.GetByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{ID: userID, IsActive: &active}, nil
				}
				teamRepo.GetByUserIDFunc = func(ctx context.Context, userID string) (*domain.Team, error) {
					return &domain.Team{Name: "team-1"}, nil
				}
				teamRepo.GetByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						Name: "team-1",
						Members: []domain.TeamMember{
							{ID: "user-1", IsActive: &active},
							{ID: "user-2", IsActive: &active},
						},
					}, nil
				}
				prRepo.CreateFunc = func(ctx context.Context, pr *domain.PullRequest) error {
					return nil
				}
			},
		},
		{
			name: "user not found",
			setupMocks: func(prRepo *MockPRRepository, teamRepo *MockTeamRepository, userRepo *MockUserRepository) {
				userRepo.GetByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return nil, repository.ErrUserNotFound
				}
			},
			expectedError: domain.NewError(domain.ErrCodeNotFound, "resource not found"),
		},
		{
			name: "PR already exists",
			setupMocks: func(prRepo *MockPRRepository, teamRepo *MockTeamRepository, userRepo *MockUserRepository) {
				active := true
				userRepo.GetByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{ID: userID, IsActive: &active}, nil
				}
				teamRepo.GetByUserIDFunc = func(ctx context.Context, userID string) (*domain.Team, error) {
					return &domain.Team{Name: "team-1"}, nil
				}
				teamRepo.GetByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						Name: "team-1",
						Members: []domain.TeamMember{
							{ID: "user-1", IsActive: &active},
						},
					}, nil
				}
				prRepo.CreateFunc = func(ctx context.Context, pr *domain.PullRequest) error {
					return repository.ErrPRAlreadyExists
				}
			},
			expectedError: domain.NewError(domain.ErrCodePRExists, "PR id already exists"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := &MockPRRepository{}
			teamRepo := &MockTeamRepository{}
			userRepo := &MockUserRepository{}
			tt.setupMocks(prRepo, teamRepo, userRepo)

			service := NewService(prRepo, teamRepo, userRepo, getTestLogger())
			result, err := service.CreatePullRequest(context.Background(), "pr-1", "Test PR", "user-1")

			if tt.expectedError != nil {
				require.Error(t, err)
				var domainErr *domain.Error
				require.True(t, errors.As(err, &domainErr))
				assert.Equal(t, tt.expectedError.Code, domainErr.Code)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestService_MergePR(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(prRepo *MockPRRepository)
		expectedError *domain.Error
	}{
		{
			name: "successful merge",
			setupMocks: func(prRepo *MockPRRepository) {
				prRepo.GetByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return &domain.PullRequest{
						ID:     prID,
						Status: domain.StatusOpen,
					}, nil
				}
				prRepo.MergeFunc = func(ctx context.Context, prID string, mergedAt time.Time) error {
					return nil
				}
			},
		},
		{
			name: "PR not found",
			setupMocks: func(prRepo *MockPRRepository) {
				prRepo.GetByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return nil, repository.ErrPRNotFound
				}
			},
			expectedError: domain.NewError(domain.ErrCodeNotFound, "resource not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := &MockPRRepository{}
			tt.setupMocks(prRepo)

			service := NewService(prRepo, &MockTeamRepository{}, &MockUserRepository{}, getTestLogger())
			result, err := service.MergePR(context.Background(), "pr-1")

			if tt.expectedError != nil {
				require.Error(t, err)
				var domainErr *domain.Error
				require.True(t, errors.As(err, &domainErr))
				assert.Equal(t, tt.expectedError.Code, domainErr.Code)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, domain.StatusMerged, result.Status)
			}
		})
	}
}

func TestService_ReassignReviewer(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(prRepo *MockPRRepository, teamRepo *MockTeamRepository)
		expectedError *domain.Error
	}{
		{
			name: "successful reassignment",
			setupMocks: func(prRepo *MockPRRepository, teamRepo *MockTeamRepository) {
				prRepo.GetByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return &domain.PullRequest{
						ID:                prID,
						Status:            domain.StatusOpen,
						AuthorID:          "author-1",
						AssignedReviewers: []string{"reviewer-1", "reviewer-2"},
					}, nil
				}
				teamRepo.GetByUserIDFunc = func(ctx context.Context, userID string) (*domain.Team, error) {
					return &domain.Team{Name: "team-1"}, nil
				}
				active := true
				teamRepo.GetByNameFunc = func(ctx context.Context, teamName string) (*domain.Team, error) {
					return &domain.Team{
						Name: "team-1",
						Members: []domain.TeamMember{
							{ID: "reviewer-1", IsActive: &active},
							{ID: "reviewer-2", IsActive: &active},
							{ID: "reviewer-3", IsActive: &active},
						},
					}, nil
				}
				prRepo.ChangeReviewerFunc = func(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
					return nil
				}
			},
		},
		{
			name: "PR already merged",
			setupMocks: func(prRepo *MockPRRepository, teamRepo *MockTeamRepository) {
				prRepo.GetByIDFunc = func(ctx context.Context, prID string) (*domain.PullRequest, error) {
					return &domain.PullRequest{
						ID:     prID,
						Status: domain.StatusMerged,
					}, nil
				}
			},
			expectedError: domain.NewError(domain.ErrCodePRMerged, "cannot reassign on merged PR"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := &MockPRRepository{}
			teamRepo := &MockTeamRepository{}
			tt.setupMocks(prRepo, teamRepo)

			service := NewService(prRepo, teamRepo, &MockUserRepository{}, getTestLogger())
			result, newReviewerID, err := service.ReassignReviewer(context.Background(), "pr-1", "reviewer-1")

			if tt.expectedError != nil {
				require.Error(t, err)
				var domainErr *domain.Error
				require.True(t, errors.As(err, &domainErr))
				assert.Equal(t, tt.expectedError.Code, domainErr.Code)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, newReviewerID)
			}
		})
	}
}
