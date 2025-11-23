package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/platonso/avito-pr-service/internal/config"
	"github.com/platonso/avito-pr-service/internal/db"
	"github.com/platonso/avito-pr-service/internal/repository/postgres"
	"github.com/platonso/avito-pr-service/internal/service/pr"
	"github.com/platonso/avito-pr-service/internal/service/stats"
	"github.com/platonso/avito-pr-service/internal/service/team"
	"github.com/platonso/avito-pr-service/internal/service/user"
	"github.com/platonso/avito-pr-service/internal/transport/handlers"
	"log/slog"
	"net/http"
	"time"
)

type App struct {
	cfg    *config.Config
	l      *slog.Logger
	dbPool *pgxpool.Pool
	server *http.Server
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

	router := a.setupRoutes()
	a.server = &http.Server{
		Addr:         ":" + a.cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

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

func (a *App) setupRoutes() *gin.Engine {
	teamRepo := postgres.NewTeamRepository(a.dbPool)
	userRepo := postgres.NewUserRepository(a.dbPool)
	prRepo := postgres.NewPRRepository(a.dbPool)

	teamService := team.NewService(teamRepo, a.l)
	userService := user.NewService(userRepo, a.l)
	prService := pr.NewService(prRepo, teamRepo, userRepo, a.l)
	statsService := stats.NewService(prRepo, a.l)

	teamHandler := handlers.NewTeamHandler(teamService, a.l)
	userHandler := handlers.NewUserHandler(userService, a.l)
	prHandler := handlers.NewPRHandler(prService, a.l)
	statsHandler := handlers.NewStatsHandler(statsService, a.l)

	router := gin.New()
	router.Use(gin.Recovery())

	teams := router.Group("/team")
	teams.POST("/add", teamHandler.CreateTeam)
	teams.GET("/get", teamHandler.GetTeam)

	users := router.Group("/users")
	users.POST("/setIsActive", userHandler.SetIsActive)
	users.GET("/getReview", userHandler.GetReview)

	pullRequest := router.Group("/pullRequest")
	pullRequest.POST("/create", prHandler.CreatePR)
	pullRequest.POST("/merge", prHandler.MergePR)
	pullRequest.POST("/reassign", prHandler.ReassignReviewer)

	stat := router.Group("/stats")
	stat.GET("/reviewers", statsHandler.GetReviewerStats)
	stat.GET("/pullRequests", statsHandler.GetPRStats)

	return router
}

func (a *App) Run() error {
	a.l.Info("starting server", slog.String("address", a.server.Addr))
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.l.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var errs []error

	// Graceful shutdown of HTTP server
	if a.server != nil {
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown server: %w", err))
		} else {
			a.l.Info("server shutdown completed")
		}
	}

	// Close DB conn
	if a.dbPool != nil {
		a.dbPool.Close()
		a.l.Info("database connections closed")
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}
