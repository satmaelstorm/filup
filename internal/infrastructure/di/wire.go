//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/satmaelstorm/filup/internal/infrastructure/config"
	"github.com/satmaelstorm/filup/internal/infrastructure/ctx"
	"github.com/satmaelstorm/filup/internal/infrastructure/logs"
	"github.com/satmaelstorm/filup/internal/infrastructure/web"
	"github.com/satmaelstorm/filup/internal/infrastructure/web/routes"
)

func InitWebServer() *web.Server {
	wire.Build(
		ctx.ProvideContext,
		config.ProvideConfig,
		logs.ProvideLoggers,
		routes.ProvideRoutes,
		web.ProvideWebServer,
	)
	return &web.Server{}
}
