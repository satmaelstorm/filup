package port

import "io"

type StorageCleaner interface {
	RemoveMeta(fileName string) error
	RemoveParts(partsNames []string) error
}

type StorageMeta interface {
	PutMetaFile(fileName string, content []byte) error
	GetMetaFile(fileName string) ([]byte, error)
}

type StoragePart interface {
	PutFilePart(fullPartName string, filesize int64, content io.Reader) error
	GetLoadedFilePartsNames(fileName string) ([]string, error)
}

type PartsComposer interface {
	ComposeFileParts(destFileName string, fullPartsName []string, tags map[string]string) (PartsComposerResult, error)
}

type PartsComposerResult interface {
	GetBucket() string
	GetName() string
	GetSize() int64
}
