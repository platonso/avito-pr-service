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

	a.initGin()

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

func (a *App) initGin() {
	a.router = gin.Default()

	a.router.Use(gin.Logger())
	a.router.Use(gin.Recovery())
}

func (a *App) Run() error {
	return a.router.Run(":" + a.cfg.HTTPPort)
}

func (a *App) Shutdown(ctx context.Context) error {
	a.dbPool.Close()
	return nil
}
