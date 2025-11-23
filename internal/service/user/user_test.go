package user

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

type MockUserRepository struct {
	SetIsActiveFunc    func(ctx context.Context, userID string, isActive bool) error
	GetByIDFunc        func(ctx context.Context, userID string) (*domain.User, error)
	GetPRsByUserIDFunc func(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
}

func (m *MockUserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	if m.SetIsActiveFunc != nil {
		return m.SetIsActiveFunc(ctx, userID, isActive)
	}
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserRepository) GetPRsByUserID(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	if m.GetPRsByUserIDFunc != nil {
		return m.GetPRsByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func getTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestService_SetUserIsActive(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		isActive       bool
		setupMocks     func(*MockUserRepository)
		expectedError  *domain.Error
		validateResult func(*testing.T, *domain.User)
	}{
		{
			name:     "successful activation",
			userID:   "user-1",
			isActive: true,
			setupMocks: func(userRepo *MockUserRepository) {
				active := true
				userRepo.SetIsActiveFunc = func(ctx context.Context, userID string, isActive bool) error {
					return nil
				}
				userRepo.GetByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{
						ID:       userID,
						Name:     "User 1",
						TeamName: "team-1",
						IsActive: &active,
					}, nil
				}
			},
			validateResult: func(t *testing.T, user *domain.User) {
				assert.NotNil(t, user)
				assert.Equal(t, "user-1", user.ID)
				assert.NotNil(t, user.IsActive)
				assert.True(t, *user.IsActive)
			},
		},
		{
			name:     "successful deactivation",
			userID:   "user-1",
			isActive: false,
			setupMocks: func(userRepo *MockUserRepository) {
				active := false
				userRepo.SetIsActiveFunc = func(ctx context.Context, userID string, isActive bool) error {
					return nil
				}
				userRepo.GetByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{
						ID:       userID,
						Name:     "User 1",
						TeamName: "team-1",
						IsActive: &active,
					}, nil
				}
			},
			validateResult: func(t *testing.T, user *domain.User) {
				assert.NotNil(t, user)
				assert.NotNil(t, user.IsActive)
				assert.False(t, *user.IsActive)
			},
		},
		{
			name:     "user not found",
			userID:   "user-1",
			isActive: true,
			setupMocks: func(userRepo *MockUserRepository) {
				userRepo.SetIsActiveFunc = func(ctx context.Context, userID string, isActive bool) error {
					return repository.ErrUserNotFound
				}
			},
			expectedError: domain.NewError(domain.ErrCodeNotFound, "resource not found"),
		},
		{
			name:     "repository error on SetIsActive",
			userID:   "user-1",
			isActive: true,
			setupMocks: func(userRepo *MockUserRepository) {
				userRepo.SetIsActiveFunc = func(ctx context.Context, userID string, isActive bool) error {
					return errors.New("database connection error")
				}
			},
			expectedError: nil,
		},
		{name: "repository error on GetByID",
			userID:   "user-1",
			isActive: true,
			setupMocks: func(userRepo *MockUserRepository) {
				userRepo.SetIsActiveFunc = func(ctx context.Context, userID string, isActive bool) error {
					return nil
				}
				userRepo.GetByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return nil, errors.New("database connection error")
				}
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &MockUserRepository{}
			tt.setupMocks(userRepo)

			service := NewService(userRepo, getTestLogger())
			result, err := service.SetUserIsActive(context.Background(), tt.userID, tt.isActive)

			switch {
			case tt.expectedError != nil:
				require.Error(t, err)
				var domainErr *domain.Error
				require.True(t, errors.As(err, &domainErr))
				assert.Equal(t, tt.expectedError.Code, domainErr.Code)
				assert.Nil(t, result)

			case tt.name == "repository error on SetIsActive" || tt.name == "repository error on GetByID":
				require.Error(t, err)
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

func TestService_GetPRsByUserID(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMocks     func(*MockUserRepository)
		expectedError  *domain.Error
		validateResult func(*testing.T, []domain.PullRequestShort)
	}{
		{
			name:   "successful get PRs",
			userID: "user-1",
			setupMocks: func(userRepo *MockUserRepository) {
				active := true
				userRepo.GetByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{
						ID:       userID,
						Name:     "User 1",
						IsActive: &active,
					}, nil
				}
				userRepo.GetPRsByUserIDFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return []domain.PullRequestShort{
						{
							ID:       "pr-1",
							Name:     "PR 1",
							AuthorID: userID,
							Status:   domain.StatusOpen,
						},
						{
							ID:       "pr-2",
							Name:     "PR 2",
							AuthorID: userID,
							Status:   domain.StatusMerged,
						},
					}, nil
				}
			},
			validateResult: func(t *testing.T, prs []domain.PullRequestShort) {
				assert.Len(t, prs, 2)
				assert.Equal(t, "pr-1", prs[0].ID)
				assert.Equal(t, domain.StatusOpen, prs[0].Status)
				assert.Equal(t, "pr-2", prs[1].ID)
				assert.Equal(t, domain.StatusMerged, prs[1].Status)
			},
		},
		{
			name:   "empty PRs list",
			userID: "user-1",
			setupMocks: func(userRepo *MockUserRepository) {
				active := true
				userRepo.GetByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{
						ID:       userID,
						Name:     "User 1",
						IsActive: &active,
					}, nil
				}
				userRepo.GetPRsByUserIDFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return []domain.PullRequestShort{}, nil
				}
			},
			validateResult: func(t *testing.T, prs []domain.PullRequestShort) {
				assert.Empty(t, prs)
			},
		},
		{
			name:   "user not found",
			userID: "user-1",
			setupMocks: func(userRepo *MockUserRepository) {
				userRepo.GetByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return nil, repository.ErrUserNotFound
				}
			},
			expectedError: domain.NewError(domain.ErrCodeNotFound, "resource not found"),
		},
		{
			name:   "repository error on GetByID",
			userID: "user-1",
			setupMocks: func(userRepo *MockUserRepository) {
				userRepo.GetByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return nil, errors.New("database connection error")
				}
			},
			expectedError: nil,
		},
		{
			name:   "repository error on GetPRsByUserID",
			userID: "user-1",
			setupMocks: func(userRepo *MockUserRepository) {
				active := true
				userRepo.GetByIDFunc = func(ctx context.Context, userID string) (*domain.User, error) {
					return &domain.User{
						ID:       userID,
						Name:     "User 1",
						IsActive: &active,
					}, nil
				}
				userRepo.GetPRsByUserIDFunc = func(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
					return nil, errors.New("database connection error")
				}
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &MockUserRepository{}
			tt.setupMocks(userRepo)

			service := NewService(userRepo, getTestLogger())
			result, err := service.GetPRsByUserID(context.Background(), tt.userID)

			switch {
			case tt.expectedError != nil:
				require.Error(t, err)
				var domainErr *domain.Error
				require.True(t, errors.As(err, &domainErr))
				assert.Equal(t, tt.expectedError.Code, domainErr.Code)
				assert.Nil(t, result)

			case tt.name == "repository error on GetByID" || tt.name == "repository error on GetPRsByUserID":
				require.Error(t, err)
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
