package pr

import (
	"context"
	"errors"
	"fmt"
	"github.com/platonso/avito-pr-service/internal/domain"
	"github.com/platonso/avito-pr-service/internal/repository"
	"log/slog"
	"math/rand"
	"time"
)

type Service struct {
	prRepo   repository.PRRepository
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
	log      *slog.Logger
}

func NewService(
	prRepo repository.PRRepository,
	teamRepo repository.TeamRepository,
	userRepo repository.UserRepository,
	log *slog.Logger,
) *Service {
	return &Service{
		prRepo:   prRepo,
		teamRepo: teamRepo,
		userRepo: userRepo,
		log:      log,
	}
}

func (s *Service) CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {
	// Check author existence
	_, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.log.Warn("PR author not found", slog.String("author_id", authorID))
			return nil, domain.NewError(domain.ErrCodeNotFound, "resource not found")
		}
	}
	// Get author's team
	team, err := s.teamRepo.GetByUserID(ctx, authorID)
	if err != nil {
		if errors.Is(err, repository.ErrTeamNotFound) {
			s.log.Warn("user not found", slog.String("user_id", authorID))
			return nil, domain.NewError(domain.ErrCodeNotFound, "resource not found")
		}
		s.log.Error(err.Error())
		return nil, fmt.Errorf("failed to get author's team: %w", err)
	}

	// Get active members (without author)
	activeMembers, err := s.getActiveTeamMembers(ctx, team.Name, authorID)
	if err != nil {
		return nil, err
	}

	// Select 0-2 reviewers
	reviewers := s.selectRandomReviewers(activeMembers, 2)

	// Create PR
	prCreatedTime := time.Now()
	pr := &domain.PullRequest{
		ID:                prID,
		Name:              prName,
		AuthorID:          authorID,
		Status:            domain.StatusOpen,
		CreatedAt:         prCreatedTime,
		AssignedReviewers: reviewers,
	}

	err = s.prRepo.Create(ctx, pr)
	if err != nil {
		if errors.Is(err, repository.ErrPRAlreadyExists) {
			s.log.Warn("PR already exists", slog.String("pr_id", prID))
			return nil, domain.NewError(domain.ErrCodePRExists, "PR id already exists")
		}
		s.log.Error(err.Error())
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	return pr, nil
}

func (s *Service) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	// Get PR with reviewers
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, repository.ErrPRNotFound) {
			s.log.Warn("PR not found", slog.String("pr_id", prID))
			return nil, domain.NewError(domain.ErrCodeNotFound, "resource not found")
		}
		s.log.Error(err.Error())
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	// Check that PR is not merged
	if pr.Status == domain.StatusMerged {
		s.log.Warn("PR already merged", slog.String("pr_id", prID))
		return pr, nil
	}

	// Merge PR
	mergeTime := time.Now()
	err = s.prRepo.Merge(ctx, prID, mergeTime)
	if err != nil {
		s.log.Error(err.Error())
		return nil, fmt.Errorf("failed to merge PR: %w", err)
	}

	pr.Status = domain.StatusMerged
	pr.MergedAt = &mergeTime

	return pr, nil
}

func (s *Service) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	// Get PR with reviewers
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, repository.ErrPRNotFound) {
			s.log.Warn("PR not found", slog.String("pr_id", prID))
			return nil, "", domain.NewError(domain.ErrCodeNotFound, "resource not found")
		}
		s.log.Error(err.Error())
		return nil, "", fmt.Errorf("failed to get PR: %w", err)
	}

	// Check that PR is not merged
	if pr.Status == domain.StatusMerged {
		s.log.Warn("cannot reassign reviewer for merged PR", slog.String("pr_id", prID))
		return nil, "", domain.NewError(domain.ErrCodePRMerged, "cannot reassign on merged PR")
	}

	// Check that old reviewer assigned to PR
	if !s.containsReviewer(pr.AssignedReviewers, oldReviewerID) {
		s.log.Warn("reviewer not assigned to PR",
			slog.String("pr_id", prID),
			slog.String("reviewer_id", oldReviewerID))
		return nil, "", domain.NewError(domain.ErrCodeNotAssigned, "reviewer is not assigned to this PR")
	}

	// Get new reviewer's team
	team, err := s.teamRepo.GetByUserID(ctx, oldReviewerID)
	if err != nil {
		s.log.Error(err.Error())
		return nil, "", fmt.Errorf("failed to get reviewer's team: %w", err)
	}

	// Get active team members (excluding author and old reviewer)
	activeMembers, err := s.getActiveTeamMembers(ctx, team.Name, oldReviewerID, pr.AuthorID)
	if err != nil {
		return nil, "", err
	}

	// Exclude assigned reviewers
	availableMembers := s.filterOutReviewers(activeMembers, pr.AssignedReviewers)

	// Choose new reviewer
	newReviewerID := s.selectRandomReviewer(availableMembers)
	if newReviewerID == "" {
		s.log.Warn("no available reviewers for reassignment",
			slog.String("pr_id", prID),
			slog.String("old_reviewer_id", oldReviewerID))
		return nil, "", domain.NewError(domain.ErrCodeNoCandidate, "no active replacement candidate in team")
	}

	// Change reviewers in DB
	err = s.prRepo.ChangeReviewer(ctx, prID, oldReviewerID, newReviewerID)
	if err != nil {
		s.log.Error(err.Error())
		return nil, "", fmt.Errorf("failed to reassign reviewer: %w", err)
	}

	// Updated PR locally
	for i, id := range pr.AssignedReviewers {
		if id == oldReviewerID {
			pr.AssignedReviewers[i] = newReviewerID
			break
		}
	}

	return pr, newReviewerID, nil
}

func (s *Service) getActiveTeamMembers(ctx context.Context, teamName string, excludeUserIDs ...string) ([]string, error) {
	team, err := s.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		return nil, err
	}

	excludeSet := make(map[string]bool)
	for _, id := range excludeUserIDs {
		excludeSet[id] = true
	}

	var activeMembers []string
	for _, member := range team.Members {
		if *member.IsActive && !excludeSet[member.ID] {
			activeMembers = append(activeMembers, member.ID)
		}
	}

	return activeMembers, nil
}

func (s *Service) selectRandomReviewers(candidates []string, maxCount int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	shuffled := make([]string, len(candidates))
	copy(shuffled, candidates)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	if len(shuffled) > maxCount {
		return shuffled[:maxCount]
	}
	return shuffled
}

func (s *Service) selectRandomReviewer(candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}
	return candidates[rand.Intn(len(candidates))]
}

func (s *Service) containsReviewer(reviewers []string, reviewerID string) bool {
	for _, id := range reviewers {
		if id == reviewerID {
			return true
		}
	}
	return false
}

// Exclude assigned reviewers
func (s *Service) filterOutReviewers(candidates []string, existingReviewers []string) []string {
	existingSet := make(map[string]bool)
	for _, id := range existingReviewers {
		existingSet[id] = true
	}

	var result []string
	for _, candidate := range candidates {
		if !existingSet[candidate] {
			result = append(result, candidate)
		}
	}
	return result
}
