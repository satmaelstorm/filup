package domain

import (
	"context"
	jsoniter "github.com/json-iterator/go"
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
	userTags      map[string]string
}

func ProvideMetaUploader(
	ctxProvider port.ContextProvider,
	config port.UploaderConfig,
	storage port.StorageMeta,
	uuidProvider UuidProvider,
	poster port.Poster,
) *MetaUploader {
	return &MetaUploader{
		uploaderCfg:  config,
		metaStorage:  storage,
		UuidProvider: uuidProvider,
		poster:       poster,
		ctx:          ctxProvider.Ctx(),
	}
}

type MetaUploader struct {
	uploaderCfg  port.UploaderConfig
	metaStorage  port.StorageMeta
	UuidProvider UuidProvider
	poster       port.Poster
	ctx          context.Context
}

func (m *MetaUploader) Handle(headers [][2]string, body []byte) ([]byte, error) {
	im, err := m.extractParams(body)
	if err != nil {
		return nil, err
	}
	if im.uuidGenerated {
		body, err = m.addUuidToBody(body, im.uuid)
		if err != nil {
			return nil, err
		}
	}

	chunks := m.prepareChunks(im)

	body, err = m.addChunksToBody(body, chunks)
	if err != nil {
		return nil, err
	}

	err = m.postBeforeUpload(headers, body)
	if err != nil {
		return nil, err
	}

	metaContent, err := m.renderMetaContent(chunks)
	if err != nil {
		return nil, err
	}

	err = m.putMetaFile(chunks.GetUUID(), metaContent)
	if err != nil {
		return nil, err
	}

	return metaContent, nil
}

func (m *MetaUploader) renderMetaContent(chunks dto.UploaderStartResult) ([]byte, error) {
	content, err := jsoniter.Marshal(chunks)
	if err != nil {
		return nil, exceptions.NewApiError(http.StatusInternalServerError, err.Error())
	}
	return content, nil
}

func (m *MetaUploader) putMetaFile(uuid string, content []byte) error {
	err := m.metaStorage.PutMetaFile(MetaFileName(uuid), content)
	if err != nil {
		return exceptions.NewApiError(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func (m *MetaUploader) postBeforeUpload(headers [][2]string, body []byte) error {
	if nil == m.uploaderCfg.GetCallbackBefore() {
		return nil
	}
	httpResult, httpCode, err := m.poster.Post(m.ctx, *m.uploaderCfg.GetCallbackBefore(), m.uploaderCfg.GetHttpTimeout(), body, headers...)
	if err != nil {
		return exceptions.NewApiError(http.StatusBadGateway, "Post error: "+err.Error())
	}
	if httpCode < 200 || httpCode > 299 {
		return exceptions.NewApiError(httpCode, string(httpResult))
	}
	return nil
}

func (m *MetaUploader) addUuidToBody(body []byte, uid string) ([]byte, error) {
	newBody, err := sjson.SetBytes(body, m.uploaderCfg.GetInfoFieldName()+".uuid", uid)
	if err != nil {
		return body, exceptions.NewApiError(http.StatusInternalServerError, err.Error())
	}
	return newBody, nil
}

func (m *MetaUploader) addChunksToBody(body []byte, chunks dto.UploaderStartResult) ([]byte, error) {
	addJson, err := jsoniter.Marshal(chunks)
	if err != nil {
		return nil, exceptions.NewApiError(http.StatusInternalServerError, err.Error())
	}
	newBody, err := sjson.SetRawBytes(body, m.uploaderCfg.GetInfoFieldName()+".chunks_info", addJson)
	if err != nil {
		return nil, exceptions.NewApiError(http.StatusInternalServerError, err.Error())
	}
	return newBody, nil
}

func (m *MetaUploader) prepareChunks(im innerMeta) dto.UploaderStartResult {
	chunkSize := m.uploaderCfg.GetChunkLength()

	if chunkSize > im.size {
		chunkFileName := ChunkFileName(im.uuid, 0)
		return dto.NewUploaderStartResult(
			im.uuid,
			map[string]dto.UploaderChunk{chunkFileName: dto.NewUploaderChunk(chunkFileName, im.size)},
			im.size,
			im.userTags,
		)
	}

	chunksCnt := im.size / chunkSize
	lastSize := im.size - (chunksCnt * chunkSize)

	chunks := make(map[string]dto.UploaderChunk, chunksCnt+1)

	for i := int64(0); i < chunksCnt; i++ {
		chunkFileName := ChunkFileName(im.uuid, int(i))
		chunks[chunkFileName] = dto.NewUploaderChunk(chunkFileName, chunkSize)
	}

	chunkFileName := ChunkFileName(im.uuid, int(chunksCnt))
	chunks[chunkFileName] = dto.NewUploaderChunk(chunkFileName, lastSize)

	return dto.NewUploaderStartResult(im.uuid, chunks, im.size, im.userTags)
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
	im.userTags = m.extractUserTags(uploaderInfo)
	return im, nil
}

func (m *MetaUploader) extractUserTags(uploaderInfo gjson.Result) map[string]string {
	result := make(map[string]string)
	tags := uploaderInfo.Get("user_tags")
	if !tags.Exists() {
		return result
	}
	tags.ForEach(func(key, value gjson.Result) bool {
		result[key.String()] = value.String()
		return true
	})
	return result
}
