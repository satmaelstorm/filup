package port

import (
	"context"
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
