package dto

type UploaderChunk struct {
	size int64
	name string
}

func (c UploaderChunk) GetSize() int64 {
	return c.size
}

func (c UploaderChunk) GetName() string {
	return c.name
}

func NewUploaderChunk(name string, size int64) UploaderChunk {
	return UploaderChunk{
		size: size,
		name: name,
	}
}

type UploaderStartResult struct {
	uuid   string
	chunks []UploaderChunk
}

func (u *UploaderStartResult) GetUUID() string {
	return u.uuid
}

func (u *UploaderStartResult) GetChunks() []UploaderChunk {
	return u.chunks
}

func NewUploaderStartResult(uuid string, chunks []UploaderChunk) UploaderStartResult {
	return UploaderStartResult{
		uuid:   uuid,
		chunks: chunks,
	}
}
