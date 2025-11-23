package app

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/platonso/avito-pr-service/internal/config"
	"github.com/platonso/avito-pr-service/internal/db"
	"github.com/platonso/avito-pr-service/internal/repository/postgres"
	"github.com/platonso/avito-pr-service/internal/service"
	"github.com/platonso/avito-pr-service/internal/transport/handlers"
	"log/slog"
)

type App struct {
	cfg    *config.Config
	l      *slog.Logger
	dbPool *pgxpool.Pool
	router *gin.Engine
}

func New(ctx context.Context, cfg *config.Config, l *slog.Logger) (*App, error) {
	a := &App{
		cfg: cfg,
		l:   l,
	}

	if err := a.initDB(ctx); err != nil {
		return nil, err
	}

	if err := a.migrateDB(); err != nil {
		return nil, err
	}

	a.setupRoutes()

	return a, nil
}

func (a *App) initDB(ctx context.Context) error {
	dbPool, err := pgxpool.New(ctx, a.cfg.GetConnStr())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	a.dbPool = dbPool
	return nil
}

func (a *App) migrateDB() error {
	if err := db.Migrate(sql.OpenDB(stdlib.GetConnector(*a.dbPool.Config().ConnConfig))); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

func (a *App) setupRoutes() {
	teamRepo := postgres.NewTeamRepository(a.dbPool)
	userRepo := postgres.NewUserRepository(a.dbPool)
	prRepo := postgres.NewPRRepository(a.dbPool)

	teamService := service.NewTeamService(teamRepo, a.l)
	userService := service.NewUserService(userRepo, a.l)
	prService := service.NewPRService(prRepo, teamRepo, userRepo, a.l)

	teamHandler := handlers.NewTeamHandler(teamService, a.l)
	userHandler := handlers.NewUserHandler(userService, a.l)
	prHandler := handlers.NewPRHandler(prService, a.l)

	a.router = gin.New()
	a.router.Use(gin.Recovery())

	team := a.router.Group("/team")
	team.POST("/add", teamHandler.CreateTeam)
	team.GET("/get", teamHandler.GetTeam)

	users := a.router.Group("/users")
	users.POST("/setIsActive", userHandler.SetIsActive)
	users.GET("/getReview", userHandler.GetReview)

	pullRequest := a.router.Group("/pullRequest")
	pullRequest.POST("/create", prHandler.CreatePR)
	pullRequest.POST("/merge", prHandler.MergePR)
	pullRequest.POST("/reassign", prHandler.ReassignReviewer)

	stats := a.router.Group("/stats")
	stats.GET("/reviewers", prHandler.GetReviewerStats)
	stats.GET("/pullRequests", prHandler.GetPRStats)

}

func (a *App) Run() error {
	return a.router.Run(":" + a.cfg.HTTPPort)
}

func (a *App) Shutdown(ctx context.Context) error {
	a.dbPool.Close()
	return nil
}
