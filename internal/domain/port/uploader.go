package port

import (
	"net/url"
	"time"
)

type UploaderConfig interface {
	GetInfoFieldName() string
	GetChunkLength() int64
	GetCallbackBefore() *url.URL
	GetCallbackAfter() *url.URL
	GetCallbackDownload() *url.URL
	GetHttpTimeout() time.Duration
	GetHttpRetries() int
	GetComposerWorkers() int
}

type UploaderConfigWithConstants interface {
	UploaderConfig
	GetMaxPartsCount() int64
	GetMaxPartSize() int64
	GetOptPartSize() int64
}
