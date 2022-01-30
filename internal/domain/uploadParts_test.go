package domain

import (
	"bytes"
	"errors"
	"github.com/satmaelstorm/filup/internal/domain/dto"
	"github.com/satmaelstorm/filup/internal/domain/exceptions"
	"github.com/satmaelstorm/filup/internal/infrastructure/config"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"testing"
)

const testMeta = `{"uuid":"31991bd9-8064-11ec-829b-e4e7494803df","size":91,"user_tags":{"0":"test"},"chunks":{"31991bd9-8064-11ec-829b-e4e7494803df_part_0":{"size":91,"name":"31991bd9-8064-11ec-829b-e4e7494803df_part_0"}}}`

type clearMock interface {
	ClearMock()
}

type fakePartsMetaStorage struct {
	willReturn []byte
	willError  error
}

func (f *fakePartsMetaStorage) ClearMock() {
	f.willReturn = nil
	f.willError = nil
}

func (f *fakePartsMetaStorage) PutMetaFile(fileName string, content []byte) error {
	return f.willError
}

func (f *fakePartsMetaStorage) GetMetaFile(fileName string) ([]byte, error) {
	return f.willReturn, f.willError
}

type fakePartsPartStorage struct {
	willReturn []string
	willError  error
}

func (f *fakePartsPartStorage) ClearMock() {
	f.willReturn = nil
	f.willError = nil
}

func (f *fakePartsPartStorage) PutFilePart(fullPartName string, filesize int64, content io.Reader) error {
	return f.willError
}

func (f *fakePartsPartStorage) GetLoadedFilePartsNames(fileName string) ([]string, error) {
	return f.willReturn, f.willError
}

type fakePartsComposerRunner struct {
	hasRun bool
}

func (f *fakePartsComposerRunner) Run(metaInfo dto.UploaderStartResult) {
	f.hasRun = true
}

func (f *fakePartsComposerRunner) ClearMock() {
	f.hasRun = false
}

type fakeReadCloser struct {
}

func (f *fakeReadCloser) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (f *fakeReadCloser) Close() error {
	return nil
}

type suiteUploadParts struct {
	suite.Suite
	up *UploadParts
}

func TestUploadParts(t *testing.T) {
	suite.Run(t, new(suiteUploadParts))
}

func (s *suiteUploadParts) SetupSuite() {
	cfg := config.Uploader{
		InfoFieldName:  "_upload_info",
		ChunkLength:    1024 * 1024 * 5,
		CallbackBefore: "http://localhost",
	}.AfterLoad()

	s.up = ProvideUploadParts(
		cfg,
		new(fakePartsPartStorage),
		new(fakePartsMetaStorage),
		new(fakePartsComposerRunner),
	)
}

func (s *suiteUploadParts) TearDownTest() {
	s.up.storageMeta.(clearMock).ClearMock()
	s.up.storage.(clearMock).ClearMock()
	s.up.partsComposer.(clearMock).ClearMock()
}

func (s *suiteUploadParts) TestExtractUuid() {
	var res, uuid string
	var err error
	var e exceptions.ApiError
	var ok bool
	uuid = UuidProvider{}.NewUuid()
	res, err = s.up.extractUuid(ChunkFileName(uuid, 0))
	s.Require().Nil(err)
	s.Equal(uuid, res)

	uuid = "870915qw-76rr-11ec-8686_e4e7494803df"
	res, err = s.up.extractUuid(ChunkFileName(uuid, 0))
	s.Require().NotNil(err)
	e, ok = err.(exceptions.ApiError)
	s.Require().True(ok)
	s.Equal(e.GetCode(), http.StatusBadRequest)

	uuid = "870915qw-76rr-11ec-8686__part_0_e4e7494803df"
	res, err = s.up.extractUuid(uuid)
	s.Require().NotNil(err)
	e, ok = err.(exceptions.ApiError)
	s.Require().True(ok)
	s.Equal(e.GetCode(), http.StatusBadRequest)
}

func (s *suiteUploadParts) TestLoadMetaSuccess() {
	s.up.storageMeta.(*fakePartsMetaStorage).willReturn = []byte(testMeta)
	r, err := s.up.loadMeta("31991bd9-8064-11ec-829b-e4e7494803df")
	s.Require().Nil(err)
	s.Equal("31991bd9-8064-11ec-829b-e4e7494803df", r.Uuid)
	s.Equal(int64(91), r.Size)
}

func (s *suiteUploadParts) TestLoadMetaFailByStorage() {
	s.up.storageMeta.(*fakePartsMetaStorage).willError = errors.New("object name cannot be empty")
	_, err := s.up.loadMeta("31991bd9-8064-11ec-829b-e4e7494803df")
	s.Require().NotNil(err)
	e, ok := err.(exceptions.ApiError)
	s.Require().True(ok)
	s.Equal(http.StatusInternalServerError, e.GetCode())
}

func (s *suiteUploadParts) TestLoadMetaFailByEmptyJson() {
	_, err := s.up.loadMeta("31991bd9-8064-11ec-829b-e4e7494803df")
	s.Require().NotNil(err)
	e, ok := err.(exceptions.ApiError)
	s.Require().True(ok)
	s.Equal(http.StatusBadRequest, e.GetCode())
}

func (s *suiteUploadParts) TestLoadMetaFailByIncorrectJson() {
	s.up.storageMeta.(*fakePartsMetaStorage).willReturn = []byte(testMeta + "}}}")
	_, err := s.up.loadMeta("31991bd9-8064-11ec-829b-e4e7494803df")
	s.Require().NotNil(err)
	e, ok := err.(exceptions.ApiError)
	s.Require().True(ok)
	s.Equal(http.StatusInternalServerError, e.GetCode())
}

func (s *suiteUploadParts) TestCheckPart() {
	uuid := "31991bd9-8064-11ec-829b-e4e7494803df"
	s.up.storageMeta.(*fakePartsMetaStorage).willReturn = []byte(testMeta)
	r, err := s.up.loadMeta(uuid)
	s.Require().Nil(err)
	err = s.up.checkPart(ChunkFileName(uuid, 0), 91, r)
	s.Nil(err)

	err = s.up.checkPart(ChunkFileName(uuid, 1), 91, r)
	s.NotNil(err)
	e, ok := err.(exceptions.ApiError)
	s.Require().True(ok)
	s.Equal(http.StatusBadRequest, e.GetCode())

	err = s.up.checkPart(ChunkFileName(uuid, 0), 90, r)
	s.NotNil(err)
	e, ok = err.(exceptions.ApiError)
	s.Require().True(ok)
	s.Equal(http.StatusBadRequest, e.GetCode())
}

func (s *suiteUploadParts) TestSavePart() {
	buf := new(bytes.Reader)
	filename := ChunkFileName("31991bd9-8064-11ec-829b-e4e7494803df", 0)
	err := s.up.savePart(filename, 91, buf)
	s.Nil(err)

	s.up.storage.(*fakePartsPartStorage).willError = errors.New("object size must be provided with disable multipart upload")
	err = s.up.savePart(filename, 91, buf)
	s.NotNil(err)
}

func (s *suiteUploadParts) TestCheckAllPartsNotComplete() {
	uuid := "31991bd9-8064-11ec-829b-e4e7494803df"
	s.up.storageMeta.(*fakePartsMetaStorage).willReturn = []byte(testMeta)
	r, err := s.up.loadMeta(uuid)
	s.Require().Nil(err)
	complete, err := s.up.checkAllParts(r)
	s.Require().Nil(err)
	s.False(complete)
}

func (s *suiteUploadParts) TestCheckAllPartsComplete() {
	uuid := "31991bd9-8064-11ec-829b-e4e7494803df"
	s.up.storageMeta.(*fakePartsMetaStorage).willReturn = []byte(testMeta)
	r, err := s.up.loadMeta(uuid)
	s.Require().Nil(err)
	s.up.storage.(*fakePartsPartStorage).willReturn = []string{ChunkFileName(uuid, 0)}
	complete, err := s.up.checkAllParts(r)
	s.Require().Nil(err)
	s.True(complete)
}

func (s *suiteUploadParts) TestCheckAllPartsError() {
	uuid := "31991bd9-8064-11ec-829b-e4e7494803df"
	s.up.storageMeta.(*fakePartsMetaStorage).willReturn = []byte(testMeta)
	r, err := s.up.loadMeta(uuid)
	s.Require().Nil(err)
	s.up.storage.(*fakePartsPartStorage).willError = errors.New("MinioS3.GetLoadedFilePartsNames")
	complete, err := s.up.checkAllParts(r)
	s.Require().NotNil(err)
	s.False(complete)
}

func (s *suiteUploadParts) TestHandleNotComlete() {
	uuid := "31991bd9-8064-11ec-829b-e4e7494803df"
	s.up.storageMeta.(*fakePartsMetaStorage).willReturn = []byte(testMeta)
	s.up.storage.(*fakePartsPartStorage).willReturn = []string{}

	complete, err := s.up.Handle(ChunkFileName(uuid, 0), 91, new(fakeReadCloser))
	s.Require().Nil(err)
	s.False(complete)
}

func (s *suiteUploadParts) TestHandleComlete() {
	uuid := "31991bd9-8064-11ec-829b-e4e7494803df"
	s.up.storageMeta.(*fakePartsMetaStorage).willReturn = []byte(testMeta)
	s.up.storage.(*fakePartsPartStorage).willReturn = []string{ChunkFileName(uuid, 0)}

	complete, err := s.up.Handle(ChunkFileName(uuid, 0), 91, new(fakeReadCloser))
	s.Require().Nil(err)
	s.True(complete)
}
