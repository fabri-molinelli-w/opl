package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fabri-molinelli-w/opl/backend/db"
	"github.com/fabri-molinelli-w/opl/backend/internal/aiclient"
	api "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/api/handlers"
	"github.com/fabri-molinelli-w/opl/backend/internal/auth"
	"github.com/fabri-molinelli-w/opl/backend/internal/config"
	dbgen "github.com/fabri-molinelli-w/opl/backend/internal/db/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/repository"
	"github.com/fabri-molinelli-w/opl/backend/internal/scheduler"
	"github.com/fabri-molinelli-w/opl/backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := run(log); err != nil {
		log.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to create database pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err := applyMigrations(ctx, pool, log); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	queries := dbgen.New(pool)
	userRepo := repository.NewUserRepository(queries)
	serverRepo := repository.NewServerRepository(queries, pool)
	serviceRepo := repository.NewServiceRepository(queries)
	tokens := auth.NewTokenIssuer(cfg.JWTSecret, cfg.JWTTTL)
	authService := service.NewAuthService(userRepo, tokens, log)
	userService := service.NewUserService(userRepo, log)
	serverService := service.NewServerService(serverRepo, serviceRepo, log)
	serviceCatalog := service.NewServiceCatalog(serviceRepo, serverRepo, log)
	documentRepo := repository.NewDocumentRepository(queries, pool)
	documentService := service.NewDocumentService(documentRepo, log)
	var opsbot service.OpsbotClient
	if cfg.AIServiceURL != "" && cfg.AIInternalToken != "" {
		opsbot = aiclient.New(cfg.AIServiceURL, cfg.AIInternalToken)
	} else {
		log.Warn("OPSBot disabled: AI_SERVICE_URL or AI_INTERNAL_TOKEN not set")
	}
	chatService := service.NewChatService(opsbot, log)
	monitoringRepo := repository.NewMonitoringRepository(queries, pool)
	monitoringService := service.NewMonitoringService(monitoringRepo, scheduler.NewChecker(), cfg.HealthcheckFailsToIncident, log)
	incidentService := service.NewIncidentService(repository.NewIncidentRepository(queries, pool), log)

	if err := authService.EnsureAdminBootstrap(ctx, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		return fmt.Errorf("failed to bootstrap admin user: %w", err)
	}

	strict := handlers.NewServer(handlers.Deps{
		DB:         pool,
		Auth:       authService,
		Users:      userService,
		Servers:    serverService,
		Services:   serviceCatalog,
		Documents:  documentService,
		Chat:       chatService,
		Monitoring: monitoringService,
		Incidents:  incidentService,
		Log:        log,
	})
	authorize := auth.Authorize(auth.NewPolicy(), tokens, userRepo)
	strictHandler := api.NewStrictHandlerWithOptions(
		strict,
		[]api.StrictMiddlewareFunc{authorize},
		handlers.NewStrictOptions(log),
	)
	api.HandlerFromMux(strictHandler, r)

	srv := &http.Server{
		Addr:        ":" + cfg.HTTPPort,
		Handler:     r,
		ReadTimeout: 5 * time.Second,
		// OPSBot (POST /api/chat) puede encadenar varias tools: el timeout de escritura
		// tiene que cubrir el timeout del cliente del ai-service (90 s).
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	schedulerDone := make(chan struct{})
	if cfg.HealthcheckEnabled {
		sched := scheduler.New(
			monitoringService.DueServices,
			func(ctx context.Context, id string) error {
				_, err := monitoringService.RunCheck(ctx, id)
				return err
			},
			monitoringService.CleanUp,
			cfg.HealthcheckWorkers, 5*time.Second, log,
		)
		go func() {
			defer close(schedulerDone)
			sched.Run(ctx)
		}()
	} else {
		close(schedulerDone)
		log.Warn("health check scheduler disabled (HEALTHCHECK_ENABLED=false)")
	}

	go func() {
		log.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	<-schedulerDone
	log.Info("server stopped")
	return nil
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) error {
	dbConn := stdlib.OpenDBFromPool(pool)
	defer func() {
		if err := dbConn.Close(); err != nil {
			log.Warn("failed to close migration connection", "error", err)
		}
	}()

	goose.SetBaseFS(db.Migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.UpContext(ctx, dbConn, "migrations"); err != nil {
		return err
	}
	log.Info("migrations applied successfully")
	return nil
}
