package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/platonso/avito-pr-service/internal/domain"
	"github.com/platonso/avito-pr-service/internal/repository"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	query := `UPDATE users SET is_active = $1 WHERE user_id = $2`
	res, err := r.db.Exec(ctx, query, isActive, userID)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}
	if res.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	var user domain.User
	query := `SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1`
	err := r.db.QueryRow(ctx, query, userID).Scan(&user.ID, &user.Name, &user.TeamName, &user.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *userRepository) GetPRsByUserID(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	query := `
	SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status 
	FROM pull_requests pr
	JOIN pr_reviewers prr ON pr.pull_request_id = prr.pr_id
	WHERE prr.reviewer_id = $1
`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs by userID: %w", err)
	}
	defer rows.Close()

	prs := make([]domain.PullRequestShort, 0)
	for rows.Next() {
		var pr domain.PullRequestShort
		err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pull request: %w", err)
		}
		prs = append(prs, pr)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating pull requests: %w", err)
	}

	return prs, nil
}
