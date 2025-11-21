package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/platonso/avito-pr-service/internal/domain"
	"github.com/platonso/avito-pr-service/internal/service"
	"github.com/platonso/avito-pr-service/internal/transport/dto"
	"log/slog"
	"net/http"
)

type TeamHandler struct {
	teamService *service.TeamService
	logger      *slog.Logger
}

func NewTeamHandler(
	teamService *service.TeamService,
	logger *slog.Logger,
) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
		logger:      logger,
	}
}

func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var team domain.Team
	if !dto.BindJSON(c, h.logger, &team) {
		return
	}

	if err := h.teamService.CreateTeam(c.Request.Context(), &team); err != nil {
		dto.WriteJSONError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"team": team,
	})
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		dto.WriteJSONError(c, h.logger, domain.NewError(domain.ErrCodeBadRequest, "team_name is required"))
		return
	}

	team, err := h.teamService.GetTeam(c.Request.Context(), teamName)
	if err != nil {
		dto.WriteJSONError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, team)
}
