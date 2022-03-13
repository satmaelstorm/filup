package handlers

import (
	"errors"
	"github.com/minio/minio-go/v7"
	"github.com/satmaelstorm/filup/internal/domain/exceptions"
	"github.com/satmaelstorm/filup/internal/domain/port"
	"github.com/satmaelstorm/filup/internal/infrastructure/logs/logsEngine"
	"github.com/valyala/fasthttp"
	"net/http"
)

const DownloadUuidParameter = "uuid"

type Handlers struct {
	logger           logsEngine.ILogger
	CoreStartUpload  port.HandlerJson
	CorePartUpload   port.HandlerMultipart
	CoreFileStreamer port.HandlerStreamer
}

func ProvideHandlers(
	logger logsEngine.ILogger,
	StartUpload port.HandlerJson,
	PartUpload port.HandlerMultipart,
	CoreFileStreamer port.HandlerStreamer,
) *Handlers {
	return &Handlers{
		logger:           logger,
		CoreStartUpload:  StartUpload,
		CorePartUpload:   PartUpload,
		CoreFileStreamer: CoreFileStreamer,
	}
}

func (h *Handlers) StartUpload(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType("application/json")
	response, err := h.CoreStartUpload.Handle(h.processHeaders(&ctx.Request.Header), ctx.Request.Body())
	if err != nil {
		h.processError(ctx, err)
		return
	}
	ctx.SetBody(response)
}

func (h *Handlers) PartUpload(ctx *fasthttp.RequestCtx) {
	mf, err := ctx.Request.MultipartForm()
	if err != nil {
		h.processError(ctx, exceptions.NewApiError(http.StatusBadRequest, err))
		return
	}
	defer ctx.Request.RemoveMultipartFormFiles()
	fileSlice := mf.File["part"]
	if len(fileSlice) < 1 {
		h.processError(ctx, exceptions.NewApiError(http.StatusBadRequest, errors.New("No part field")))
		return
	}
	if fileSlice[0] == nil {
		h.processError(ctx, exceptions.NewApiError(http.StatusBadRequest, errors.New("No part field")))
		return
	}
	file, err := fileSlice[0].Open()
	if err != nil {
		h.processError(ctx, exceptions.NewApiError(http.StatusBadRequest, err))
		return
	}
	b, err := h.CorePartUpload.Handle(fileSlice[0].Filename, fileSlice[0].Size, file)
	if err != nil {
		h.processError(ctx, err)
		return
	}
	if !b {
		ctx.SetStatusCode(http.StatusContinue)
	}
	ctx.SetStatusCode(http.StatusNoContent)
}

func (h *Handlers) DownloadFile(ctx *fasthttp.RequestCtx) {
	uuid := ctx.UserValue(DownloadUuidParameter)
	fileName, ok := uuid.(string)
	if !ok {
		ctx.Response.SetStatusCode(http.StatusBadRequest)
		ctx.Response.SetBodyString("Invalid file name")
		return
	}
	streamer, info, err := h.CoreFileStreamer.GetStreamer(fileName)
	if err != nil {
		h.processError(ctx, err)
		return
	}
	ctx.Response.Header.SetContentType(info.GetContentType())
	ctx.Response.SetBodyStreamWriter(streamer)
}

func (h *Handlers) processError(ctx *fasthttp.RequestCtx, err error) {
	h.logger.Error().Println(err)
	apiErr, ok := err.(port.HttpError)
	if ok {
		code, msg := h.getBaseErrorCodeAndMsg(apiErr.GetErr(), apiErr.GetCode(), apiErr.Error())
		ctx.SetStatusCode(code)
		if code >= http.StatusInternalServerError {
			ctx.Response.SetBodyString("Internal server error")
		} else {
			ctx.SetBodyString(msg)
		}
	} else {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.Response.SetBodyString("Internal server error")
	}
}

func (h *Handlers) processHeaders(header *fasthttp.RequestHeader) [][2]string {
	result := make([][2]string, header.Len())
	ptr := 0
	header.VisitAll(func(key, value []byte) {
		k := make([]byte, len(key))
		v := make([]byte, len(value))
		copy(k, key)
		copy(v, value)
		result[ptr] = [2]string{string(k), string(v)}
		ptr += 1
	})
	return result
}

func (h *Handlers) getBaseError(err error) error {
	r := err
	for nr := errors.Unwrap(r); nr != nil; nr = errors.Unwrap(r) {
		r = nr
	}
	return r
}

func (h *Handlers) getBaseErrorCodeAndMsg(err error, defCode int, defMsg string) (int, string) {
	baseError := h.getBaseError(err)
	if baseError != nil {
		switch baseError.(type) {
		case minio.ErrorResponse:
			return baseError.(minio.ErrorResponse).StatusCode, baseError.(minio.ErrorResponse).Message
		}
	}
	return defCode, defMsg
}
