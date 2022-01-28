package port

import "github.com/satmaelstorm/filup/internal/domain/dto"

type PartComposerRunner interface {
	Run(metaInfo dto.UploaderStartResult)
}
