package domain

import (
	"context"
	"errors"
	"github.com/satmaelstorm/filup/internal/domain/exceptions"
	"github.com/satmaelstorm/filup/internal/infrastructure/config"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	"net/http"
	"net/url"
	"testing"
	"time"
)

const (
	uploadMetaTestJson1 = `
{
  "_upload_info": {
    "file_size": 52428800,
    "user_tags": {
      "tag1": "str1",
      "tag2": "str2"
    }
  },
  "other_info": {}
}
`

	uploadMetaTestJson2 = `
{
  "_upload_info": {
    "file_size": 0,
    "user_tags": {
      "tag1": "str1",
      "tag2": "str2"
    }
  }
}
`

	uploadMetaTestJson3 = `
{
  "_upload_info": {
    "file_size": 52428800,
    "user_tags": [
      "str1",
      "str2"
    ]
  }
}
`

	uploadMetaTestJson4 = `
{
  "_upload_info": {
    "file_size": 52428800,
    "user_tags": {
      "tag1": ["str1", "str3"],
      "tag2": {"str2": "str4", "str5": "str6"}
    }
  }
}
`

	uploadMetaTestJson5 = `
{
  "upload_info": {
    "file_size": 52428800,
    "user_tags": {
      "tag1": ["str1", "str3"],
      "tag2": {"str2": "str4", "str5": "str6"}
    }
  }
}
`
)

type fakeStorage struct {
	lastFilename string
}

func (f *fakeStorage) PutMetaFile(fileName string, content []byte) error {
	f.lastFilename = fileName
	return nil
}

func (f *fakeStorage) GetMetaFile(fileName string) ([]byte, error) {
	return nil, nil
}

type fakePoster struct {
	retErr  error
	retCode int
}

func (f fakePoster) Post(ctx context.Context, serviceUrl url.URL, timeOut time.Duration, body []byte, headers ...[2]string) ([]byte, int, error) {
	return nil, f.retCode, f.retErr
}

type suiteUploadMeta struct {
	suite.Suite
	uploader                MetaUploader
	uploaderWithoutCallback MetaUploader
}

func TestUploadMeta(t *testing.T) {
	suite.Run(t, new(suiteUploadMeta))
}

func (s *suiteUploadMeta) SetupSuite() {
	cfg := config.Uploader{
		InfoFieldName:  "_upload_info",
		ChunkLength:    1024 * 1024 * 5,
		CallbackBefore: "http://localhost",
	}.AfterLoad()

	s.uploader = MetaUploader{
		uploaderCfg:  cfg,
		metaStorage:  &fakeStorage{},
		poster:       fakePoster{},
		UuidProvider: ProvideUuidProvider(),
	}

	cfg2 := config.Uploader{
		InfoFieldName: "_upload_info",
		ChunkLength:   1024 * 1024 * 5,
	}.AfterLoad()

	s.uploaderWithoutCallback = MetaUploader{
		uploaderCfg:  cfg2,
		UuidProvider: ProvideUuidProvider(),
		poster:       fakePoster{},
	}
}

func (s *suiteUploadMeta) TestExtractTags() {
	ui := gjson.GetBytes([]byte(uploadMetaTestJson1), s.uploader.uploaderCfg.GetInfoFieldName())
	r := s.uploader.extractUserTags(ui)
	s.Require().NotNil(r)
	s.Require().Contains(r, "tag1")
	s.Require().Contains(r, "tag2")
	s.Equal("str1", r["tag1"])
	s.Equal("str2", r["tag2"])

	ui = gjson.GetBytes([]byte(uploadMetaTestJson3), s.uploader.uploaderCfg.GetInfoFieldName())
	r = s.uploader.extractUserTags(ui)
	s.Require().NotNil(r)
	s.Require().Contains(r, "0")
	s.Require().Contains(r, "1")
	s.Equal("str1", r["0"])
	s.Equal("str2", r["1"])

	ui = gjson.GetBytes([]byte(uploadMetaTestJson4), s.uploader.uploaderCfg.GetInfoFieldName())
	r = s.uploader.extractUserTags(ui)
	s.Require().NotNil(r)
	s.Require().Contains(r, "tag1")
	s.Require().Contains(r, "tag2")
	s.Equal("[\"str1\", \"str3\"]", r["tag1"])
	s.Equal("{\"str2\": \"str4\", \"str5\": \"str6\"}", r["tag2"])

	ui = gjson.GetBytes([]byte(uploadMetaTestJson5), s.uploader.uploaderCfg.GetInfoFieldName())
	r = s.uploader.extractUserTags(ui)
	s.Require().Equal(0, len(r))
}

func (s *suiteUploadMeta) TestExtractParams() {
	r, err := s.uploader.extractParams([]byte(uploadMetaTestJson1))
	s.Require().Nil(err)
	s.Equal(int64(52428800), r.size)
	s.Equal(2, len(r.userTags))

	r, err = s.uploader.extractParams([]byte(uploadMetaTestJson2))
	s.Require().NotNil(err)
	apiError, ok := err.(exceptions.ApiError)
	s.Require().True(ok)
	s.Equal(http.StatusBadRequest, apiError.GetCode())
	s.Equal(0, len(r.userTags))

	r, err = s.uploader.extractParams([]byte(uploadMetaTestJson5))
	s.Require().NotNil(err)
	apiError, ok = err.(exceptions.ApiError)
	s.Require().True(ok)
	s.Equal(http.StatusBadRequest, apiError.GetCode())
	s.Equal(0, len(r.userTags))
}

func (s *suiteUploadMeta) TestAddUuid() {
	uid := s.uploader.UuidProvider.NewUuid()
	r, err := s.uploader.addUuidToBody([]byte(uploadMetaTestJson1), uid)
	s.Require().Nil(err)
	s.Require().NotNil(r)
	uidFromJson := gjson.GetBytes(r, s.uploader.uploaderCfg.GetInfoFieldName()+".uuid")
	s.Require().True(uidFromJson.Exists())
	s.Equal(uid, uidFromJson.String())
}

func (s *suiteUploadMeta) TestPrepareChunks() {
	uid := s.uploader.UuidProvider.NewUuid()
	im := innerMeta{
		uuid:     uid,
		size:     1024 * 1024 * 4,
		userTags: map[string]string{"tag1": "val1"},
	}
	result := s.uploader.prepareChunks(im)
	chunkFileName0 := ChunkFileName(im.uuid, 0)
	chunkFileName7 := ChunkFileName(im.uuid, 7)
	s.Require().Equal(1, len(result.GetChunks()))
	s.Equal(uid, result.GetUUID())
	s.Require().Contains(result.GetChunks(), chunkFileName0)
	s.Equal(int64(1024*1024*4), result.GetChunks()[chunkFileName0].GetSize())
	s.Equal(ChunkFileName(uid, 0), result.GetChunks()[chunkFileName0].GetName())
	s.Equal(int64(1024*1024*4), result.GetSize())
	s.Require().Equal(len(im.userTags), len(result.GetUserTags()))
	s.Require().Contains(result.GetUserTags(), "tag1")
	s.Equal(result.GetUserTags()["tag1"], "val1")

	im.size = 1024 * 1024 * 36
	result = s.uploader.prepareChunks(im)
	s.Require().Equal(8, len(result.GetChunks()))
	s.Equal(uid, result.GetUUID())
	s.Require().Contains(result.GetChunks(), chunkFileName0)
	s.Require().Contains(result.GetChunks(), chunkFileName7)
	s.Equal(s.uploader.uploaderCfg.GetChunkLength(), result.GetChunks()[chunkFileName0].GetSize())
	s.Equal(int64(1024*1024), result.GetChunks()[chunkFileName7].GetSize())
	s.Equal(im.size, result.GetSize())
}

func (s *suiteUploadMeta) TestAddChunks() {
	uid := s.uploader.UuidProvider.NewUuid()
	im := innerMeta{uuid: uid, size: 1024 * 1024 * 36}
	result := s.uploader.prepareChunks(im)
	r, err := s.uploader.addChunksToBody([]byte(uploadMetaTestJson1), result)
	s.Require().Nil(err)
	s.Require().NotNil(r)
	chunksInfo := gjson.GetBytes(r, s.uploader.uploaderCfg.GetInfoFieldName()+".chunks_info")
	s.Require().True(chunksInfo.Exists())
	m := chunksInfo.Map()
	s.Require().Contains(m, "uuid")
	s.Require().Contains(m, "chunks")
	s.Equal(m["uuid"].String(), uid)
	s.Equal(len(result.GetChunks()), len(m["chunks"].Map()))
}

func (s *suiteUploadMeta) TestPutMetaFile() {
	im := innerMeta{
		uuid:     s.uploader.UuidProvider.NewUuid(),
		size:     1024 * 1024 * 4,
		userTags: map[string]string{"tag1": "val1"},
	}
	result := s.uploader.prepareChunks(im)
	metaContent, err := s.uploader.renderMetaContent(result)
	s.Require().NotNil(metaContent)
	s.Require().Nil(err)

	err = s.uploader.putMetaFile(result.GetUUID(), metaContent)
	s.Require().Nil(err)
	fs := s.uploader.metaStorage.(*fakeStorage)
	s.Equal(MetaFileName(im.uuid), fs.lastFilename)
}

func (s *suiteUploadMeta) TestPostBeforeHook() {
	fp := fakePoster{
		retErr:  nil,
		retCode: 200,
	}
	s.uploader.poster = fp

	var err error

	err = s.uploader.postBeforeUpload([][2]string{{"API-KEY", "qwerty"}}, []byte(uploadMetaTestJson1))
	s.Require().Nil(err)

	rErr := errors.New("test error")
	fp = fakePoster{
		retErr:  rErr,
		retCode: 200,
	}
	s.uploader.poster = fp
	err = s.uploader.postBeforeUpload([][2]string{{"API-KEY", "qwerty"}}, []byte(uploadMetaTestJson1))
	s.Require().NotNil(err)
	apiErr, ok := err.(exceptions.ApiError)
	s.Require().True(ok)
	s.Equal("Post error: "+rErr.Error(), apiErr.Error())

	fp = fakePoster{
		retErr:  nil,
		retCode: 404,
	}
	s.uploader.poster = fp
	err = s.uploader.postBeforeUpload([][2]string{{"API-KEY", "qwerty"}}, []byte(uploadMetaTestJson1))
	s.Require().NotNil(err)
	apiErr, ok = err.(exceptions.ApiError)
	s.Require().True(ok)
	s.Equal(fp.retCode, apiErr.GetCode())

	s.uploader.poster = fakePoster{}
}

func (s *suiteUploadMeta) TestPostBeforeHookNil() {
	fp := fakePoster{
		retErr:  nil,
		retCode: 404,
	}
	s.uploaderWithoutCallback.poster = fp

	err := s.uploaderWithoutCallback.postBeforeUpload([][2]string{{"API-KEY", "qwerty"}}, []byte(uploadMetaTestJson1))
	s.Require().Nil(err)

	s.uploaderWithoutCallback.poster = fakePoster{}
}
