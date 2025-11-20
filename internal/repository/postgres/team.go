package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/platonso/avito-pr-service/internal/domain"
	"github.com/platonso/avito-pr-service/internal/repository"
)

type teamRepository struct {
	db *pgxpool.Pool
}

func NewTeamRepository(db *pgxpool.Pool) repository.TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) CreateWithMembers(ctx context.Context, team *domain.Team) (err error) {
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

	// Create team
	teamQuery := `INSERT INTO teams(team_name) VALUES ($1)`
	_, err = tx.Exec(ctx, teamQuery, team.Name)
	if err != nil {
		if isDuplicateTeamKeyError(err) {
			return repository.ErrTeamAlreadyExists
		}
		return fmt.Errorf("failed to create team: %w", err)
	}

	// Create/update users
	usersQuery := `
		INSERT INTO users (user_id, username, team_name, is_active) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id)
		DO UPDATE SET
			username = $2,
			team_name = $3,
			is_active = $4
`
	for _, member := range team.Members {
		_, err = tx.Exec(ctx, usersQuery, member.ID, member.Name, team.Name, member.IsActive)
		if err != nil {
			return fmt.Errorf("failed to create/update user: %w", err)
		}
	}
	return nil
}

func (r *teamRepository) GetByName(ctx context.Context, teamName string) (*domain.Team, error) {
	// Check team existence
	exists, err := r.Exists(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to check team existence: %w", err)
	}
	if !exists {
		return nil, repository.ErrTeamNotFound
	}

	var team domain.Team
	query := `
		SELECT user_id, username, is_active 
		FROM users
		WHERE team_name = $1
`
	rows, err := r.db.Query(ctx, query, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}
	defer rows.Close()

	// Get team members
	for rows.Next() {
		var tm domain.TeamMember
		err := rows.Scan(&tm.ID, &tm.Name, &tm.IsActive)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}
		team.Members = append(team.Members, tm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team member: %w", err)
	}

	return &team, nil
}

func (r *teamRepository) GetByUserID(ctx context.Context, userID string) (*domain.Team, error) {
	// Get user's team name
	var teamName string
	query := `SELECT team_name FROM users WHERE user_id = $1`
	err := r.db.QueryRow(ctx, query, userID).Scan(&teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get user's team: %w", err)
	}

	// Get team
	return r.GetByName(ctx, teamName)
}

func (r *teamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`
	err := r.db.QueryRow(ctx, query, teamName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check team existence: %w", err)
	}
	return exists, nil
}

func isDuplicateTeamKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
