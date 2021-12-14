package port

type UploaderConfig interface {
	GetCopyHeaders() []string
	GetMaxFileLength() int64
}
