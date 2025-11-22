package domain

import "time"

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type User struct {
	ID       string `json:"user_id" binding:"required,min=1"`
	Name     string `json:"username" binding:"required,min=1"`
	TeamName string `json:"team_name" binding:"required,min=1"`
	IsActive bool   `json:"is_active" binding:"required,min=1"`
}

type Team struct {
	Name    string       `json:"team_name" binding:"required,min=1"`
	Members []TeamMember `json:"members" binding:"required,min=1,dive"`
}

type TeamMember struct {
	ID       string `json:"user_id" binding:"required,min=1"`
	Name     string `json:"username" binding:"required,min=1"`
	IsActive bool   `json:"is_active" binding:"required,min=1"`
}

type PullRequest struct {
	ID                string     `json:"pull_request_id" binding:"required,min=1"`
	Name              string     `json:"pull_request_name" binding:"required,min=1"`
	AuthorID          string     `json:"author_id" binding:"required,min=1"`
	Status            PRStatus   `json:"status" binding:"required"`
	AssignedReviewers []string   `json:"assigned_reviewers" binding:"required"`
	CreatedAt         time.Time  `json:"createdAt" binding:"required"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	ID       string   `json:"pull_request_id"`
	Name     string   `json:"pull_request_name"`
	AuthorID string   `json:"author_id"`
	Status   PRStatus `json:"status"`
}
