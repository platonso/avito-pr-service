package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/platonso/avito-pr-service/internal/domain"
	"github.com/platonso/avito-pr-service/internal/repository"
	"time"
)

type prRepository struct {
	db *pgxpool.Pool
}

func NewPRRepository(db *pgxpool.Pool) repository.PRRepository {
	return &prRepository{db: db}
}

func (r *prRepository) Create(ctx context.Context, pr *domain.PullRequest) (err error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		var e error
		if err == nil {
			e = tx.Commit(ctx)
		} else {
			e = tx.Rollback(ctx)
		}

		if err == nil && e != nil {
			err = fmt.Errorf("finishing transaction: %w", e)
		}
	}()

	// Create pull request
	prQuery := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at) 
		VALUES ($1, $2, $3, $4, $5)
`
	_, err = tx.Exec(ctx, prQuery, pr.ID, pr.Name, pr.AuthorID, pr.Status, pr.CreatedAt)
	if err != nil {
		if isDuplicatePRKeyError(err) {
			return repository.ErrPRAlreadyExists
		}
		return fmt.Errorf("failed to create PR: %w", err)
	}

	// Create pull request with reviewers
	prReviewersQuery := `INSERT INTO pr_reviewers (pr_id, reviewer_id) VALUES ($1, $2)`
	for _, reviewerID := range pr.AssignedReviewers {
		_, err = tx.Exec(ctx, prReviewersQuery, pr.ID, reviewerID)
		if err != nil {
			return fmt.Errorf("failed to assign reviewer: %w", err)
		}
	}

	return nil
}

func (r *prRepository) Merge(ctx context.Context, prID string, mergedAt time.Time) error {
	// Update merge status and date
	query := `
		UPDATE pull_requests 
		SET status = $1, merged_at = COALESCE(merged_at, $2)
		WHERE pull_request_id = $3 AND status != $1
`
	res, err := r.db.Exec(ctx, query, string(domain.StatusMerged), mergedAt, prID)
	if err != nil {
		return fmt.Errorf("failed to update merge status and date: %w", err)
	}

	if res.RowsAffected() == 0 {
		// Check pr existing
		exists, err := r.Exists(ctx, prID)
		if err != nil {
			return err
		}
		if !exists {
			return repository.ErrPRNotFound
		}
	}
	return nil
}

func (r *prRepository) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	var pr domain.PullRequest
	query := `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at 
		FROM pull_requests
		WHERE pull_request_id = $1
`
	err := r.db.QueryRow(ctx, query, prID).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrPRNotFound
		}
		return nil, fmt.Errorf("failed to get pull request by ID: %w", err)
	}

	// Get reviewers
	reviewers, err := r.GetReviewersIDs(ctx, prID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *prRepository) GetReviewersIDs(ctx context.Context, prID string) ([]string, error) {
	query := `SELECT reviewer_id FROM pr_reviewers WHERE pr_id = $1`
	rows, err := r.db.Query(ctx, query, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewers: %w", err)
	}
	defer rows.Close()

	reviewersIDs := make([]string, 0)
	for rows.Next() {
		var reviewerID string
		err := rows.Scan(&reviewerID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reviewerID: %w", err)
		}
		reviewersIDs = append(reviewersIDs, reviewerID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reviewersIDs: %w", err)
	}

	return reviewersIDs, nil
}

func (r *prRepository) ChangeReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {

	query := `
			UPDATE pr_reviewers 
			SET reviewer_id = $1 
			WHERE reviewer_id = $2 AND pr_id = $3`
	res, err := r.db.Exec(ctx, query, newReviewerID, oldReviewerID, prID)
	if err != nil {
		return fmt.Errorf("failed to update reviewer: %w", err)
	}

	if res.RowsAffected() == 0 {
		return repository.ErrPRNotFound
	}

	return nil
}

func (r *prRepository) Exists(ctx context.Context, prID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`
	err := r.db.QueryRow(ctx, query, prID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check pull request existence: %w", err)
	}
	return exists, nil
}

func isDuplicatePRKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
