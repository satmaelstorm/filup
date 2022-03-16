package routes

import "github.com/satmaelstorm/filup/internal/infrastructure/web/handlers"

const (
	Metrics = "/metrics"

	Upload                = "/upload"
	StartUpload           = Upload + "/start"
	UploadPart            = Upload + "/part"
	Download              = "/download"
	DownloadUuidParameter = handlers.DownloadUuidParameter
	DownloadFile          = Download + "/{" + DownloadUuidParameter + ":^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$}"
)
