package domain

import (
	"github.com/satmaelstorm/filup/internal/domain/dto"
	"github.com/satmaelstorm/filup/internal/domain/exceptions"
	"github.com/satmaelstorm/filup/internal/domain/port"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"net/http"
)

type innerMeta struct {
	size          int64
	uuid          string
	uuidGenerated bool
}

type MetaUploader struct {
	uploaderCfg  port.UploaderConfig
	metaStorage  port.StorageMeta
	UuidProvider UuidProvider
}

func (m *MetaUploader) Handler(body []byte) (dto.UploaderStartResult, error) {
	im, err := m.extractParams(body)
	if err != nil {
		return dto.UploaderStartResult{}, err
	}
	if im.uuidGenerated {
		body, err = m.addUuidToBody(body, im.uuid)
		if err != nil {
			return dto.UploaderStartResult{}, err
		}
	}
	return m.prepareChunks(im.uuid, im.size), nil
}

func (m *MetaUploader) addUuidToBody(body []byte, uid string) ([]byte, error) {
	newBody, err := sjson.SetBytes(body, m.uploaderCfg.GetInfoFieldName()+".uuid", uid)
	if err != nil {
		return body, exceptions.NewApiError(http.StatusInternalServerError, err.Error())
	}
	return newBody, nil
}

func (m *MetaUploader) prepareChunks(uniqId string, size int64) dto.UploaderStartResult {

	chunkSize := m.uploaderCfg.GetChunkLength()

	if chunkSize > size {
		return dto.NewUploaderStartResult(uniqId, []dto.UploaderChunk{dto.NewUploaderChunk(ChunkFileName(uniqId, 0), size)})
	}

	chunksCnt := size / chunkSize
	lastSize := size - (chunksCnt * chunkSize)

	chunks := make([]dto.UploaderChunk, chunksCnt+1)

	for i := int64(0); i < chunksCnt; i++ {
		chunks[i] = dto.NewUploaderChunk(ChunkFileName(uniqId, int(i)), chunkSize)
	}

	chunks[chunksCnt] = dto.NewUploaderChunk(ChunkFileName(uniqId, int(chunksCnt)), lastSize)

	return dto.NewUploaderStartResult(uniqId, chunks)
}

func (m *MetaUploader) extractParams(body []byte) (innerMeta, error) {
	im := innerMeta{}
	uploaderInfo := gjson.GetBytes(body, m.uploaderCfg.GetInfoFieldName())
	if !uploaderInfo.Exists() {
		return im, exceptions.NewApiError(http.StatusBadRequest, "field "+m.uploaderCfg.GetInfoFieldName()+" is required!")
	}
	fs := uploaderInfo.Get("file_size")
	if !fs.Exists() {
		return im, exceptions.NewApiError(http.StatusBadRequest, "field "+m.uploaderCfg.GetInfoFieldName()+".file_size is required!")
	}
	fileSize := fs.Int()
	if fileSize < 1 {
		return im, exceptions.NewApiError(http.StatusBadRequest, "field "+m.uploaderCfg.GetInfoFieldName()+".file_size must be greater than 0")
	}
	im.size = fileSize
	uid := uploaderInfo.Get("uuid")
	if uid.Exists() {
		im.uuid = uid.String()
	} else {
		im.uuid = m.UuidProvider.NewUuid()
		im.uuidGenerated = true
	}
	return im, nil
}

func (m *MetaUploader) ExtractUserTags(body []byte) (map[string]string, error) {
	uploaderInfo := gjson.GetBytes(body, m.uploaderCfg.GetInfoFieldName())
	if !uploaderInfo.Exists() {
		return nil, exceptions.NewApiError(http.StatusBadRequest, "field "+m.uploaderCfg.GetInfoFieldName()+" is required!")
	}
	tags := uploaderInfo.Get("user_tags")
	if !tags.Exists() {
		return nil, nil
	}
	result := make(map[string]string)
	tags.ForEach(func(key, value gjson.Result) bool {
		result[key.String()] = value.String()
		return true
	})
	return result, nil
}
