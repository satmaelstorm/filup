package handlers

import (
	"github.com/satmaelstorm/filup/internal/domain/port"
	"github.com/valyala/fasthttp"
)

type Handlers struct {
	CoreStartUpload port.HandlerJson
}

func ProvideHandlers(
	StartUpload port.HandlerJson,
) *Handlers {
	return &Handlers{
		CoreStartUpload: StartUpload,
	}
}

func (h *Handlers) StartUpload(ctx *fasthttp.RequestCtx) {

}
