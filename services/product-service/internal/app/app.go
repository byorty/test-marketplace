package app

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/byorty/test-marketplace/services/product-service/internal/auth"
	"github.com/byorty/test-marketplace/services/product-service/internal/config"
	"github.com/byorty/test-marketplace/services/product-service/internal/database"
	httptransport "github.com/byorty/test-marketplace/services/product-service/internal/handler/transport/http"
	"github.com/byorty/test-marketplace/services/product-service/internal/logger"
	"github.com/byorty/test-marketplace/services/product-service/internal/rbac"
	"github.com/byorty/test-marketplace/services/product-service/internal/repository/postgres"
	"github.com/byorty/test-marketplace/services/product-service/internal/service"
	"gorm.io/gorm"
)

type App struct {
	cfg *config.Config
	log *slog.Logger
	db *gorm.DB
	server *http.Server
	sqlDB *sql.DB
}

func New() (*App, error) {
	app := &App{}

	app.cfg = config.MustLoad()

	log, err := logger.New(app.cfg.Log)
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}
	app.log = log

	db, err := database.New(app.cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	app.db = db

	repository := postgres.New(app.db)

	productService := service.New(app.log, repository)

	handler := httptransport.New(productService, app.log)

	jwtManager := auth.New(
    app.cfg.JWT.Secret,
    app.cfg.JWT.Issuer,
	)

	authorizer := rbac.New()

	router := httptransport.NewRouter(handler, jwtManager, authorizer)

	app.server = &http.Server{
		Addr: fmt.Sprintf("%s:%d", app.cfg.HTTP.Host, app.cfg.HTTP.Port),
		Handler: router,
	}

	return app, nil
}