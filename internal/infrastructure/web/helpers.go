package web

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	"github.com/satmaelstorm/filup/internal/infrastructure/config"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var ErrHttpError = errors.New("http error")

type RequestFunc func(url.URL, time.Duration, string, *bytes.Reader, ...[2]string) ([]byte, int, error)

type RequestHelpers struct {
	requestFunc RequestFunc
}

func ProvideRequestHelpers() *RequestHelpers {
	return &RequestHelpers{requestFunc: doRequest}
}

func (rh *RequestHelpers) SetRequestFunc(f RequestFunc) {
	rh.requestFunc = f
}

type response struct {
	result []byte
	err    error
	code   int
}

func (rh *RequestHelpers) Get(
	ctx context.Context,
	serviceUrl url.URL,
	timeOut time.Duration,
	headers ...[2]string,
) ([]byte, int, error) {
	chResult := make(chan response)
	go func() {
		httpResult, httpCode, err := rh.requestFunc(
			serviceUrl,
			timeOut,
			http.MethodGet,
			nil,
			headers...,
		)
		chResult <- response{
			result: httpResult,
			err:    err,
			code:   httpCode,
		}
	}()
	select {
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	case result := <-chResult:
		return result.result, result.code, result.err
	}
}

func (rh *RequestHelpers) Post(
	ctx context.Context,
	serviceUrl url.URL,
	timeOut time.Duration,
	body []byte,
	headers ...[2]string,
) ([]byte, int, error) {
	chResult := make(chan response)
	go func() {
		httpResult, httpCode, err := rh.requestFunc(
			serviceUrl,
			timeOut,
			http.MethodPost,
			bytes.NewReader(body),
			headers...,
		)
		chResult <- response{
			result: httpResult,
			err:    err,
			code:   httpCode,
		}
	}()
	select {
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	case result := <-chResult:
		return result.result, result.code, result.err
	}
}

func doRequest(
	url url.URL,
	timeOut time.Duration,
	method string,
	body *bytes.Reader,
	headers ...[2]string,
) ([]byte, int, error) {
	req, err := http.NewRequest(method, url.String(), body)
	if err != nil {
		return nil, 0, err
	}
	if len(headers) > 0 {
		for _, h := range headers {
			req.Header.Set(h[0], h[1])
		}
	}
	req.Header.Set("User-Agent", config.ProjectHttpClientName)
	client := &http.Client{Timeout: timeOut * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 299 {
		return nil,
			resp.StatusCode,
			errors.Wrap(ErrHttpError, "status code is "+strconv.Itoa(resp.StatusCode))
	}
	result, err := io.ReadAll(resp.Body)
	return result, resp.StatusCode, err
}
