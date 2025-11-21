package dto

import (
	"github.com/platonso/avito-pr-service/internal/domain"
)

// User request DTO
type SetIsActiveReq struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive *bool  `json:"is_active" binding:"required"`
}
type GetUserReviewsResp struct {
	UserID       string                    `json:"user_id"`
	PullRequests []domain.PullRequestShort `json:"pull_requests"`
}

// Pull request (request DTO)
type CreatePRReq struct {
	PRID     string `json:"pull_request_id" binding:"required"`
	PRName   string `json:"pull_request_name" binding:"required"`
	AuthorID string `json:"author_id" binding:"required"`
}

type MergePRReq struct {
	PRID string `json:"pull_request_id" binding:"required"`
}

type ReassignPRReq struct {
	PRID          string `json:"pull_request_id" binding:"required"`
	OldReviewerID string `json:"old_reviewer_id" binding:"required"`
}
