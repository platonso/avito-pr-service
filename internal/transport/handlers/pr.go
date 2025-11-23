package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/platonso/avito-pr-service/internal/service/pr"
	"github.com/platonso/avito-pr-service/internal/transport/dto"
	"log/slog"
	"net/http"
)

type PRHandler struct {
	prService *pr.Service
	logger    *slog.Logger
}

func NewPRHandler(
	prService *pr.Service,
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

	pullRequest, err := h.prService.CreatePullRequest(c.Request.Context(), req.PRID, req.PRName, req.AuthorID)
	if err != nil {
		dto.WriteJSONError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusCreated, dto.PRResp{PR: pullRequest})
}

func (h *PRHandler) MergePR(c *gin.Context) {
	var req dto.MergePRReq
	if !dto.BindJSON(c, h.logger, &req) {
		return
	}

	pullRequest, err := h.prService.MergePR(c.Request.Context(), req.PRID)
	if err != nil {
		dto.WriteJSONError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, dto.PRResp{PR: pullRequest})
}

func (h *PRHandler) ReassignReviewer(c *gin.Context) {
	var req dto.ReassignPRReq
	if !dto.BindJSON(c, h.logger, &req) {
		return
	}

	pullRequest, newReviewerID, err := h.prService.ReassignReviewer(c.Request.Context(), req.PRID, req.OldReviewerID)
	if err != nil {
		dto.WriteJSONError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, dto.ReassignPRResp{
		PR:         pullRequest,
		ReplacedBy: newReviewerID,
	})
}
