package app

import (
	"context"
	"net/http"

	"github.com/byorty/test-marketplace/services/common/auth"
	"github.com/byorty/test-marketplace/services/common/rbac"
	"github.com/byorty/test-marketplace/services/product-service/internal/config"
	httptransport "github.com/byorty/test-marketplace/services/product-service/internal/transport"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func RunServer(
	
	lifecycle fx.Lifecycle,

	handler *httptransport.ProductHandler,

	cfg *config.Config,

	log *zap.Logger,

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