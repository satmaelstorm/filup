package domain

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/satmaelstorm/filup/internal/domain/dto"
	"github.com/satmaelstorm/filup/internal/domain/exceptions"
	"github.com/satmaelstorm/filup/internal/domain/port"
	"io"
	"net/http"
	"strconv"
)

type UploadParts struct {
	config      port.UploaderConfig
	storage     port.StoragePart
	storageMeta port.StorageMeta

	partsComposer *PartsComposer
}

func ProvideUploadParts(
	cfg port.UploaderConfig,
	storage port.StoragePart,
	storageMeta port.StorageMeta,
	composer *PartsComposer,
) *UploadParts {
	up := new(UploadParts)
	up.config = cfg
	up.storage = storage
	up.storageMeta = storageMeta
	up.partsComposer = composer
	return up
}

func (up *UploadParts) Handle(filename string, size int64, file io.ReadCloser) (isComplete bool, err error) {
	defer func() {
		_ = file.Close()
	}()

	uuid, err := up.extractUuid(filename)
	if err != nil {
		return false, err
	}
	metaInfo, err := up.loadMeta(uuid)
	if err != nil {
		return false, err
	}

	if err := up.checkPart(filename, size, metaInfo); err != nil {
		return false, err
	}

	if err := up.savePart(filename, size, file); err != nil {
		return false, err
	}

	done, err := up.checkAllParts(metaInfo)
	if err != nil {
		return false, err
	}

	if done {
		up.partsComposer.Run(metaInfo)
	}
	return done, nil
}

func (up *UploadParts) extractUuid(filename string) (string, error) {
	uuid, err := ExtractUuidFromPartName(filename)
	if err != nil {
		return "", err
	}
	if !IsCorrectUuid(uuid) {
		return "", exceptions.NewApiError(http.StatusBadRequest, "incorrect part name - must start with uuid")
	}
	return uuid, nil
}

func (up *UploadParts) loadMeta(uuid string) (dto.UploaderStartResult, error) {
	//TODO add inmemory cache
	metaInfoBytes, err := up.storageMeta.GetMetaFile(MetaFileName(uuid))
	if err != nil { //TODO process known errors to StatusBadRequest
		return dto.UploaderStartResult{}, exceptions.NewApiError(http.StatusInternalServerError, "error in meta storage: "+err.Error())
	}
	if len(metaInfoBytes) < 1 {
		return dto.UploaderStartResult{}, exceptions.NewApiError(http.StatusBadRequest, "incorrect part file name: upload not started")
	}
	var metaInfo dto.UploaderStartResult
	err = jsoniter.Unmarshal(metaInfoBytes, &metaInfo)
	if err != nil {
		return dto.UploaderStartResult{}, exceptions.NewApiError(http.StatusInternalServerError, "error while deserialize meta: "+err.Error())
	}
	return metaInfo, nil
}

func (up *UploadParts) checkPart(filename string, filesize int64, metaInfo dto.UploaderStartResult) error {
	part, ok := metaInfo.GetChunks()[filename]
	if !ok {
		return exceptions.NewApiError(http.StatusBadRequest, "incorrect part file name: no part with this name in upload")
	}
	if filesize != part.GetSize() {
		return exceptions.NewApiError(http.StatusBadRequest,
			"incorrect part: incorrect size, must be "+strconv.Itoa(int(part.GetSize()))+" bytes but got "+strconv.Itoa(int(filesize))+" bytes")
	}
	return nil
}

func (up *UploadParts) savePart(filename string, filesize int64, file io.Reader) error {
	err := up.storage.PutFilePart(filename, filesize, file)
	if err != nil {
		return exceptions.NewApiError(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func (up *UploadParts) checkAllParts(metaInfo dto.UploaderStartResult) (bool, error) {
	list, err := up.storage.GetLoadedFilePartsNames(metaInfo.GetUUID())
	if err != nil {
		return false, exceptions.NewApiError(http.StatusInternalServerError, err.Error())
	}
	mapList := make(map[string]bool, len(list))
	for _, fn := range list {
		mapList[fn] = true
	}
	parts := metaInfo.GetChunks()
	foundParts := 0
	for partName, _ := range parts {
		if _, ok := mapList[partName]; ok {
			foundParts += 1
		}
	}
	if foundParts >= len(parts) {
		return true, nil
	}
	return false, nil
}
