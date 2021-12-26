package routes

import (
	"github.com/fasthttp/router"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/satmaelstorm/filup/internal/infrastructure/config"
	"github.com/satmaelstorm/filup/internal/infrastructure/logs/logsEngine"
	"github.com/satmaelstorm/filup/internal/infrastructure/metrics"
	"github.com/satmaelstorm/filup/internal/infrastructure/web/handlers"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"github.com/valyala/fasthttp/pprofhandler"
	"runtime/debug"
	"strconv"
	"time"
)

var (
	requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: config.ProjectMetricsNamespace,
		Subsystem: "webserver",
		Name:      "request_duration",
		Buckets:   metrics.StdHttpBuckets,
	}, []string{"path", "code"})
	requestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: config.ProjectMetricsNamespace,
		Subsystem: "webserver",
		Name:      "request_count",
	}, []string{"code"})
)

func ProvideRoutes(hs *handlers.Handlers, logger logsEngine.Loggers) *router.Router {
	prometheus.MustRegister(requestCount, requestDuration)
	r := router.New()

	r.PanicHandler = func(ctx *fasthttp.RequestCtx, i interface{}) {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Response.SetBodyString("Internal server error")
		requestCount.WithLabelValues(strconv.Itoa(fasthttp.StatusInternalServerError)).Inc()
		logger.Critical().Printf("%s : %s\n", i, string(debug.Stack()))
	}
	r.GET("/debug/pprof/{ep:*}", pprofhandler.PprofHandler)
	r.GET(Metrics, fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler()))

	innerHandler := getDomainRouter(hs).Handler

	r.ANY("/{path:*}", func(ctx *fasthttp.RequestCtx) {
		timeStart := time.Now()

		innerHandler(ctx)

		timeElapsed := time.Since(timeStart)
		c := ctx.Response.StatusCode()
		code := strconv.Itoa(c)
		requestCount.WithLabelValues(code).Inc()
		if c != fasthttp.StatusNotFound &&
			c != fasthttp.StatusMethodNotAllowed &&
			c != fasthttp.StatusForbidden &&
			c != fasthttp.StatusUnauthorized &&
			c != fasthttp.StatusNotAcceptable {
			requestDuration.WithLabelValues(ctx.UserValue("path").(string), code).Observe(timeElapsed.Seconds())
		}
	})

	return r
}

func getDomainRouter(hs *handlers.Handlers) *router.Router {
	r := router.New()
	r.GET("/upload", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.SetBodyString("upload api")
	})

	r.POST(StartUpload, hs.StartUpload)

	return r
}
