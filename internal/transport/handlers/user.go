package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/platonso/avito-pr-service/internal/domain"
	"github.com/platonso/avito-pr-service/internal/service"
	"github.com/platonso/avito-pr-service/internal/transport/dto"
	"log/slog"
	"net/http"
)

type UserHandler struct {
	userService *service.UserService
	logger      *slog.Logger
}

func NewUserHandler(
	userService *service.UserService,
	logger *slog.Logger,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *UserHandler) SetIsActive(c *gin.Context) {
	var req dto.SetIsActiveReq
	if !dto.BindJSON(c, h.logger, &req) {
		return
	}

	user, err := h.userService.SetUserIsActive(c.Request.Context(), req.UserID, *req.IsActive)
	if err != nil {
		dto.WriteJSONError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, dto.SetIsActiveResp{User: user})
}

func (h *UserHandler) GetReview(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		err := domain.NewError(domain.ErrCodeBadRequest, "user_id is required")
		dto.WriteJSONError(c, h.logger, err)
		return
	}

	prs, err := h.userService.GetPRsByUserID(c.Request.Context(), userID)
	if err != nil {
		dto.WriteJSONError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, dto.GetUserReviewsResp{
		UserID:       userID,
		PullRequests: prs,
	})
}
