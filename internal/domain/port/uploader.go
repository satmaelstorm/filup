package port

type UploaderConfig interface {
	GetInfoFieldName() string
	GetChunkLength() int64
}

type UploaderConfigWithConstants interface {
	UploaderConfig
	GetMaxPartsCount() int64
	GetMaxPartSize() int64
	GetOptPartSize() int64
}
