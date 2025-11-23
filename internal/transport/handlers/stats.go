package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/platonso/avito-pr-service/internal/service/stats"
	"github.com/platonso/avito-pr-service/internal/transport/dto"
	"log/slog"
	"net/http"
)

type StatsHandler struct {
	statsService *stats.Service
	logger       *slog.Logger
}

func NewStatsHandler(
	statsService *stats.Service,
	logger *slog.Logger,
) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
		logger:       logger,
	}
}

func (h *StatsHandler) GetReviewerStats(c *gin.Context) {
	s, err := h.statsService.GetReviewerAssignmentsStats(c.Request.Context())
	if err != nil {
		dto.WriteJSONError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, dto.ReviewerStatsResp{Stats: s})
}

func (h *StatsHandler) GetPRStats(c *gin.Context) {
	s, err := h.statsService.GetPRStats(c.Request.Context())
	if err != nil {
		dto.WriteJSONError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, dto.PRStatsResp{Stats: s})
}
