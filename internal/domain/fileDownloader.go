package domain

import (
	"bufio"
	"context"
	"github.com/pkg/errors"
	"github.com/satmaelstorm/filup/internal/domain/exceptions"
	"github.com/satmaelstorm/filup/internal/domain/port"
	"io"
	"net/http"
)

type FileDownloader struct {
	streamer port.FileStreamer
	logger   port.Logger
	config   port.UploaderConfig
	getter   port.Getter
	ctx      context.Context
}

func ProvideFileDownloader(
	ctxProvider port.ContextProvider,
	config port.UploaderConfig,
	streamer port.FileStreamer,
	getter port.Getter,
	logger port.Logger,
) *FileDownloader {
	return &FileDownloader{
		streamer: streamer,
		logger:   logger,
		config:   config,
		getter:   getter,
		ctx:      ctxProvider.Ctx(),
	}
}

func (fd *FileDownloader) GetStreamer(headers [][2]string, fileName string) (func(writer *bufio.Writer), port.FileInfo, error) {
	if fd.config.GetCallbackDownload() != nil {
		httpResult, httpCode, err := fd.getter.Get(fd.ctx, *fd.config.GetCallbackDownload(), fd.config.GetHttpTimeout(), headers...)
		if err != nil {
			return nil, nil, exceptions.NewApiError(http.StatusBadGateway, errors.Wrap(err, "Get error"))
		}
		if httpCode < 200 || httpCode > 299 {
			return nil, nil, exceptions.NewApiError(httpCode, errors.New(string(httpResult)))
		}
	}
	stream, info, err := fd.streamer.GetFileStream(fileName)
	if err != nil {
		return nil, nil, exceptions.NewApiError(http.StatusInternalServerError, err)
	}
	return fd.getStreamerFunc(stream), info, nil
}

func (fd *FileDownloader) getStreamerFunc(stream io.ReadCloser) func(writer *bufio.Writer) {
	return func(writer *bufio.Writer) {
		defer func(stream io.Closer) {
			err := stream.Close()
			if err != nil {
				fd.logger.Error().Println(err)
			}
		}(stream)
		_, err := writer.ReadFrom(stream)
		if err != nil {
			fd.logger.Error().Println(err)
		}
	}
}
