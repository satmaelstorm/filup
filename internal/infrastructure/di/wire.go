//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/satmaelstorm/filup/internal/domain"
	"github.com/satmaelstorm/filup/internal/domain/port"
	"github.com/satmaelstorm/filup/internal/infrastructure/appctx"
	"github.com/satmaelstorm/filup/internal/infrastructure/config"
	"github.com/satmaelstorm/filup/internal/infrastructure/logs"
	"github.com/satmaelstorm/filup/internal/infrastructure/logs/logsEngine"
	"github.com/satmaelstorm/filup/internal/infrastructure/storage"
	"github.com/satmaelstorm/filup/internal/infrastructure/web"
	"github.com/satmaelstorm/filup/internal/infrastructure/web/handlers"
	"github.com/satmaelstorm/filup/internal/infrastructure/web/routes"
)

func InitWebServer() (*web.Server, error) {
	wire.Build(
		wire.Bind(new(port.ContextProvider), new(*appctx.CoreContext)),
		wire.Bind(new(port.StorageMeta), new(*storage.MinioS3)),
		wire.Bind(new(port.StoragePart), new(*storage.MinioS3)),
		wire.Bind(new(port.PartsComposer), new(*storage.MinioS3)),
		wire.Bind(new(port.Poster), new(*web.RequestHelpers)),
		wire.Bind(new(port.HandlerJson), new(*domain.MetaUploader)),
		wire.Bind(new(port.HandlerMultipart), new(*domain.UploadParts)),
		wire.Bind(new(port.CriticalLogger), new(*logsEngine.Loggers)),
		wire.Bind(new(logsEngine.ILogger), new(*logsEngine.Loggers)),
		wire.Bind(new(port.PartComposerRunner), new(*domain.PartsComposer)),

		appctx.ProvideContext,
		config.ProvideConfig,
		config.ProvideUploaderConfig,
		logs.ProvideLoggers,
		routes.ProvideRoutes,
		web.ProvideWebServer,
		handlers.ProvideHandlers,
		web.ProvideRequestHelpers,
		storage.ProvideMinioS3,
		domain.ProvideMetaUploader,
		domain.ProvideUuidProvider,
		domain.ProvideUploadParts,
		domain.ProvidePartsComposer,
	)
	return &web.Server{}, nil
}
