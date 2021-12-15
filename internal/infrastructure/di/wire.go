//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/satmaelstorm/filup/internal/infrastructure/appctx"
	"github.com/satmaelstorm/filup/internal/infrastructure/config"
	"github.com/satmaelstorm/filup/internal/infrastructure/logs"
	"github.com/satmaelstorm/filup/internal/infrastructure/web"
	"github.com/satmaelstorm/filup/internal/infrastructure/web/routes"
)

func InitWebServer() *web.Server {
	wire.Build(
		appctx.ProvideContext,
		config.ProvideConfig,
		logs.ProvideLoggers,
		routes.ProvideRoutes,
		web.ProvideWebServer,
	)
	return &web.Server{}
}
