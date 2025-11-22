package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/platonso/avito-pr-service/internal/service"
	"github.com/platonso/avito-pr-service/internal/transport/dto"
	"log/slog"
	"net/http"
)

type PRHandler struct {
	prService *service.PRService
	logger    *slog.Logger
}

func NewPRHandler(
	prService *service.PRService,
	logger *slog.Logger,
) *PRHandler {
	return &PRHandler{
		prService: prService,
		logger:    logger,
	}
}

func (h *PRHandler) CreatePR(c *gin.Context) {
	var req dto.CreatePRReq
	if !dto.BindJSON(c, h.logger, &req) {
		return
	}

	pr, err := h.prService.CreatePullRequest(c.Request.Context(), req.PRID, req.PRName, req.AuthorID)
	if err != nil {
		dto.WriteJSONError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusCreated, dto.PRResp{PR: pr})
}

func (h *PRHandler) MergePR(c *gin.Context) {
	var req dto.MergePRReq
	if !dto.BindJSON(c, h.logger, &req) {
		return
	}

	pr, err := h.prService.MergePR(c.Request.Context(), req.PRID)
	if err != nil {
		dto.WriteJSONError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, dto.PRResp{PR: pr})
}

func (h *PRHandler) ReassignReviewer(c *gin.Context) {
	var req dto.ReassignPRReq
	if !dto.BindJSON(c, h.logger, &req) {
		return
	}

	pr, newReviewerID, err := h.prService.ReassignReviewer(c.Request.Context(), req.PRID, req.OldReviewerID)
	if err != nil {
		dto.WriteJSONError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, dto.ReassignPRResp{
		PR:         pr,
		ReplacedBy: newReviewerID,
	})
}

func (h *PRHandler) GetReviewerStats(c *gin.Context) {
	stats, err := h.prService.GetReviewerAssignmentsStats(c.Request.Context())
	if err != nil {
		dto.WriteJSONError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, dto.ReviewerStatsResp{Stats: stats})
}

func (h *PRHandler) GetPRStats(c *gin.Context) {
	stats, err := h.prService.GetPRStats(c.Request.Context())
	if err != nil {
		dto.WriteJSONError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, dto.PRStatsResp{Stats: stats})
}
