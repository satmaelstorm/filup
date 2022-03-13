package port

import (
	"bufio"
	"context"
	"io"
	"net/url"
	"time"
)

type Getter interface {
	Get(
		ctx context.Context,
		serviceUrl url.URL,
		timeOut time.Duration,
		headers ...[2]string,
	) ([]byte, int, error)
}

type Poster interface {
	Post(
		ctx context.Context,
		serviceUrl url.URL,
		timeOut time.Duration,
		body []byte,
		headers ...[2]string,
	) ([]byte, int, error)
}

type HttpError interface {
	error
	GetCode() int
	GetErr() error
}

type HandlerJson interface {
	Handle(headers [][2]string, body []byte) ([]byte, error)
}

type HandlerMultipart interface {
	Handle(string, int64, io.ReadCloser) (bool, error)
}

type HandlerStreamer interface {
	GetStreamer(fileName string) (func(writer *bufio.Writer), FileInfo, error)
}
