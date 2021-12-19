package domain

import (
	"github.com/satmaelstorm/filup/internal/domain/exceptions"
	"github.com/satmaelstorm/filup/internal/infrastructure/config"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	"net/http"
	"testing"
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
}

func (f fakeStorage) PutMetaFile(fileName string, content []byte) error {
	return nil
}

func (f fakeStorage) GetMetaFile(fileName string) ([]byte, error) {
	return nil, nil
}

type suiteUploadMeta struct {
	suite.Suite
	uploader MetaUploader
}

func TestUploadMeta(t *testing.T) {
	suite.Run(t, new(suiteUploadMeta))
}

func (s *suiteUploadMeta) SetupSuite() {
	s.uploader = MetaUploader{
		uploaderCfg: config.Uploader{
			InfoFieldName: "_upload_info",
			ChunkLength:   1024 * 1025 * 5,
		},
		metaStorage: fakeStorage{},
	}
}

func (s *suiteUploadMeta) TestExtractTags() {
	r, err := s.uploader.ExtractUserTags([]byte(uploadMetaTestJson1))
	s.Require().Nil(err)
	s.Require().NotNil(r)
	s.Require().Contains(r, "tag1")
	s.Require().Contains(r, "tag2")
	s.Equal("str1", r["tag1"])
	s.Equal("str2", r["tag2"])

	r, err = s.uploader.ExtractUserTags([]byte(uploadMetaTestJson3))
	s.Require().Nil(err)
	s.Require().NotNil(r)
	s.Require().Contains(r, "0")
	s.Require().Contains(r, "1")
	s.Equal("str1", r["0"])
	s.Equal("str2", r["1"])

	r, err = s.uploader.ExtractUserTags([]byte(uploadMetaTestJson4))
	s.Require().Nil(err)
	s.Require().NotNil(r)
	s.Require().Contains(r, "tag1")
	s.Require().Contains(r, "tag2")
	s.Equal("[\"str1\", \"str3\"]", r["tag1"])
	s.Equal("{\"str2\": \"str4\", \"str5\": \"str6\"}", r["tag2"])

	r, err = s.uploader.ExtractUserTags([]byte(uploadMetaTestJson5))
	s.Require().NotNil(err)
	s.Require().Nil(r)
}

func (s *suiteUploadMeta) TestExtractParams() {
	r, err := s.uploader.extractParams([]byte(uploadMetaTestJson1))
	s.Require().Nil(err)
	s.Equal(int64(52428800), r.size)

	r, err = s.uploader.extractParams([]byte(uploadMetaTestJson2))
	s.Require().NotNil(err)
	apiError, ok := err.(exceptions.ApiError)
	s.Require().True(ok)
	s.Equal(http.StatusBadRequest, apiError.GetCode())

	r, err = s.uploader.extractParams([]byte(uploadMetaTestJson5))
	s.Require().NotNil(err)
	apiError, ok = err.(exceptions.ApiError)
	s.Require().True(ok)
	s.Equal(http.StatusBadRequest, apiError.GetCode())
}

func (s *suiteUploadMeta) TestAddUuid() {
	uid := UuidProvider{}.NewUuid()
	r, err := s.uploader.addUuidToBody([]byte(uploadMetaTestJson1), uid)
	s.Require().Nil(err)
	s.Require().NotNil(r)
	uidFromJson := gjson.GetBytes(r, s.uploader.uploaderCfg.GetInfoFieldName()+".uuid")
	s.Require().True(uidFromJson.Exists())
	s.Equal(uid, uidFromJson.String())
}
