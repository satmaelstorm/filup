package dto

type UploaderChunk struct {
	size int
	name string
}

func (c UploaderChunk) GetSize() int {
	return c.size
}

func (c UploaderChunk) GetName() string {
	return c.name
}

type UploaderStartResult struct {
	uuid   string
	chunks []UploaderChunk
}

func (u UploaderStartResult) GetUUID() string {
	return u.uuid
}

func (u UploaderStartResult) GetChunks() []UploaderChunk {
	return u.chunks
}
