package port

type StorageMeta interface {
	PutMetaFile(fileName string, content []byte) error
	GetMetaFile(fileName string) ([]byte, error)
}

type StoragePart interface {
	PutFilePart(fullPartName string, content []byte) error
	GetLoadedFilePartsNames(fileName string) ([]string, error)
}

type PartsComposer interface {
	ComposeFileParts(destFileName string, fullPartsName []string) (PartsComposerResult, error)
}

type PartsComposerResult interface {
	GetBucket() string
	GetName() string
	GetSize() int64
}
