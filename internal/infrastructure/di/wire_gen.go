// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package di

import (
	"github.com/satmaelstorm/filup/internal/domain"
	"github.com/satmaelstorm/filup/internal/infrastructure/appctx"
	"github.com/satmaelstorm/filup/internal/infrastructure/cache"
	"github.com/satmaelstorm/filup/internal/infrastructure/config"
	"github.com/satmaelstorm/filup/internal/infrastructure/logs"
	"github.com/satmaelstorm/filup/internal/infrastructure/storage"
	"github.com/satmaelstorm/filup/internal/infrastructure/web"
	"github.com/satmaelstorm/filup/internal/infrastructure/web/handlers"
	"github.com/satmaelstorm/filup/internal/infrastructure/web/routes"
)

// Injectors from wire.go:

func InitWebServer() (*web.Server, error) {
	coreContext := appctx.ProvideContext()
	configuration := config.ProvideConfig()
	loggers := logs.ProvideLoggers(configuration)
	uploaderConfig := config.ProvideUploaderConfig()
	cacheCache, err := cache.ProvideMetaCache(configuration, loggers)
	if err != nil {
		return nil, err
	}
	minioS3, err := storage.ProvideMinioS3(configuration, coreContext, cacheCache)
	if err != nil {
		return nil, err
	}
	uuidProvider := domain.ProvideUuidProvider()
	requestHelpers := web.ProvideRequestHelpers()
	metaUploader := domain.ProvideMetaUploader(coreContext, uploaderConfig, minioS3, uuidProvider, requestHelpers)
	partsComposer := domain.ProvidePartsComposer(coreContext, minioS3, minioS3, uploaderConfig, loggers, requestHelpers)
	uploadParts := domain.ProvideUploadParts(uploaderConfig, minioS3, minioS3, partsComposer)
	fileDownloader := domain.ProvideFileDownloader(minioS3, loggers)
	handlersHandlers := handlers.ProvideHandlers(loggers, metaUploader, uploadParts, fileDownloader)
	router := routes.ProvideRoutes(handlersHandlers, loggers)
	server := web.ProvideWebServer(coreContext, router, configuration, loggers)
	return server, nil
}
