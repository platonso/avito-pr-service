package dto

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/platonso/avito-pr-service/internal/domain"
	"log/slog"
	"net/http"
)

// User response DTO
type SetIsActiveResp struct {
	User *domain.User `json:"user"`
}

// Pull request response DTO
type PRResp struct {
	PR *domain.PullRequest `json:"pr"`
}

type ReassignPRResp struct {
	PR         *domain.PullRequest `json:"pr"`
	ReplacedBy string              `json:"replaced_by"`
}

// Statistics response DTO
type ReviewerStatsResp struct {
	Stats []domain.ReviewerStat `json:"stats"`
}

type PRStatsResp struct {
	Stats []domain.PullRequestStat `json:"stats"`
}

// Error response DTO
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func WriteJSONError(c *gin.Context, logger *slog.Logger, err error) {
	var domainErr *domain.Error
	if errors.As(err, &domainErr) {
		statusCode := http.StatusBadRequest

		switch domainErr.Code {
		case domain.ErrCodeNotFound:
			statusCode = http.StatusNotFound
		case domain.ErrCodeTeamExists,
			domain.ErrCodePRExists,
			domain.ErrCodePRMerged,
			domain.ErrCodeNotAssigned,
			domain.ErrCodeNoCandidate:
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, ErrorResponse{
			Error: struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}{
				Code:    string(domainErr.Code),
				Message: domainErr.Message,
			},
		})
		return
	}

	logger.Error("internal server error", "error", err)
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		},
	})
}

func BindJSON(c *gin.Context, logger *slog.Logger, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		WriteJSONError(c, logger, domain.NewError(domain.ErrCodeBadRequest, err.Error()))
		return false
	}
	return true
}
