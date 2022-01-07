package domain

import "github.com/satmaelstorm/filup/internal/domain/port"

type UploadParts struct {
	config  port.UploaderConfig
	storage port.StoragePart
}
