package domain

import (
	"bufio"
	"github.com/satmaelstorm/filup/internal/domain/exceptions"
	"github.com/satmaelstorm/filup/internal/domain/port"
	"io"
	"net/http"
)

type FileDownloader struct {
	streamer port.FileStreamer
	logger   port.Logger
}

func ProvideFileDownloader(
	streamer port.FileStreamer,
	logger port.Logger,
) *FileDownloader {
	return &FileDownloader{
		streamer: streamer,
		logger:   logger,
	}
}

func (fd *FileDownloader) GetStreamer(fileName string) (func(writer *bufio.Writer), port.FileInfo, error) {
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
