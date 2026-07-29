package main

import (
	"github.com/byorty/test-marketplace/services/product-service/internal/app"
	"go.uber.org/fx"
)

func main() {
	fx.New(app.Module).Run()
}
