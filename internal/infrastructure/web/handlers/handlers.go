package handlers

import (
	"github.com/satmaelstorm/filup/internal/domain/exceptions"
	"github.com/satmaelstorm/filup/internal/domain/port"
	"github.com/satmaelstorm/filup/internal/infrastructure/logs/logsEngine"
	"github.com/valyala/fasthttp"
	"net/http"
)

type Handlers struct {
	logger          logsEngine.ILogger
	CoreStartUpload port.HandlerJson
	CorePartUpload  port.HandlerMultipart
}

func ProvideHandlers(
	logger logsEngine.ILogger,
	StartUpload port.HandlerJson,
	PartUpload port.HandlerMultipart,
) *Handlers {
	return &Handlers{
		logger:          logger,
		CoreStartUpload: StartUpload,
		CorePartUpload:  PartUpload,
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
		h.processError(ctx, exceptions.NewApiError(http.StatusBadRequest, err.Error()))
		return
	}
	defer ctx.Request.RemoveMultipartFormFiles()
	fileSlice := mf.File["part"]
	if len(fileSlice) < 1 {
		h.processError(ctx, exceptions.NewApiError(http.StatusBadRequest, "No part field"))
		return
	}
	if fileSlice[0] == nil {
		h.processError(ctx, exceptions.NewApiError(http.StatusBadRequest, "No part field"))
		return
	}
	file, err := fileSlice[0].Open()
	if err != nil {
		h.processError(ctx, exceptions.NewApiError(http.StatusBadRequest, err.Error()))
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

func (h *Handlers) processError(ctx *fasthttp.RequestCtx, err error) {
	h.logger.Error().Println(err)
	apiErr, ok := err.(port.HttpError)
	if ok {
		code := apiErr.GetCode()
		ctx.SetStatusCode(code)
		if code >= http.StatusInternalServerError {
			ctx.Request.SetBodyString("Internal server error")
		} else {
			ctx.SetBodyString(apiErr.Error())
		}
	} else {
		ctx.SetStatusCode(http.StatusInternalServerError)
		ctx.Request.SetBodyString("Internal server error")
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
