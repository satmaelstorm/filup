package port

type UploaderConfig interface {
	GetInfoFieldName() string
	GetChunkLength() int64
}
