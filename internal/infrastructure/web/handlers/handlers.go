package handlers

import (
	"github.com/satmaelstorm/filup/internal/domain/port"
	"github.com/satmaelstorm/filup/internal/infrastructure/logs/logsEngine"
	"github.com/valyala/fasthttp"
	"net/http"
)

type Handlers struct {
	logger          logsEngine.ILogger
	CoreStartUpload port.HandlerJson
}

func ProvideHandlers(
	logger logsEngine.ILogger,
	StartUpload port.HandlerJson,
) *Handlers {
	return &Handlers{
		logger:          logger,
		CoreStartUpload: StartUpload,
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

func (h *Handlers) processError(ctx *fasthttp.RequestCtx, err error) {
	h.logger.Error().Println(err)
	apiErr, ok := err.(port.HttpError)
	if ok {
		ctx.SetStatusCode(apiErr.GetCode())
		ctx.SetBodyString(apiErr.Error())
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
