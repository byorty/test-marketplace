package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/byorty/test-marketplace/services/product-service/internal/auth"
	"github.com/byorty/test-marketplace/services/product-service/internal/config"
	httptransport "github.com/byorty/test-marketplace/services/product-service/internal/handler/transport/http"
	"github.com/byorty/test-marketplace/services/product-service/internal/rbac"
	"go.uber.org/fx"
)

func RunServer(
	
	lifecycle fx.Lifecycle,

	handler *httptransport.Handler,

	cfg *config.Config,

	log *slog.Logger,

	jwt *auth.Validator,

	authorizer *rbac.Authorizer,
	
)	{
	
	router := httptransport.NewRouter(handler, jwt, authorizer)

	server := &http.Server{
		Addr: cfg.HTTP.Address(),
		Handler: router,
	}

	lifecycle.Append(fx.Hook{

		OnStart: func(ctx context.Context) error {
			go server.ListenAndServe()
			log.Info("http server started")
			return nil
		},

		OnStop: func(ctx context.Context) error {
			log.Info("http server stopped")
			return server.Shutdown(ctx)
		},
	})

}

var ServerModule = fx.Invoke(RunServer)