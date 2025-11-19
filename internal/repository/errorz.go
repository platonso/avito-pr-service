package repository

import "errors"

var (
	ErrTeamAlreadyExists = errors.New("team name already exists")
	ErrTeamNotFound      = errors.New("team not found")

	ErrUserNotFound = errors.New("user not found")

	ErrPRAlreadyExists = errors.New("PR id already exists")
	ErrPRNotFound      = errors.New("PR not found")
)
