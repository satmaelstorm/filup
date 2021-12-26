package dto

type UploaderChunk struct {
	Size int64  `json:"size"`
	Name string `json:"name"`
}

func (c UploaderChunk) GetSize() int64 {
	return c.Size
}

func (c UploaderChunk) GetName() string {
	return c.Name
}

func NewUploaderChunk(name string, size int64) UploaderChunk {
	return UploaderChunk{
		Size: size,
		Name: name,
	}
}

type UploaderStartResult struct {
	Uuid     string            `json:"uuid"`
	Size     int64             `json:"size"`
	UserTags map[string]string `json:"user_tags"`
	Chunks   []UploaderChunk   `json:"chunks"`
}

func (u *UploaderStartResult) GetUUID() string {
	return u.Uuid
}

func (u *UploaderStartResult) GetChunks() []UploaderChunk {
	return u.Chunks
}

func (u *UploaderStartResult) GetSize() int64 {
	return u.Size
}

func (u *UploaderStartResult) GetUserTags() map[string]string {
	return u.UserTags
}

func NewUploaderStartResult(
	uuid string,
	chunks []UploaderChunk,
	size int64,
	userTags map[string]string,
) UploaderStartResult {
	return UploaderStartResult{
		Uuid:     uuid,
		Chunks:   chunks,
		Size:     size,
		UserTags: userTags,
	}
}
