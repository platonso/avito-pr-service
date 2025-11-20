package domain

import "time"

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type User struct {
	ID       string `json:"user_id"`
	Name     string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	Name    string       `json:"team_name"`
	Members []TeamMember `json:"members"`
}

type TeamMember struct {
	ID       string `json:"user_id"`
	Name     string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type PullRequest struct {
	ID                string     `json:"pull_request_id"`
	Name              string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            PRStatus   `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         time.Time  `json:"createdAt"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	ID       string   `json:"pull_request_id"`
	Name     string   `json:"pull_request_name"`
	AuthorID string   `json:"author_id"`
	Status   PRStatus `json:"status"`
}
