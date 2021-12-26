package web

import (
	"context"
	"github.com/fasthttp/router"
	"github.com/satmaelstorm/filup/internal/infrastructure/appctx"
	"github.com/satmaelstorm/filup/internal/infrastructure/config"
	"github.com/satmaelstorm/filup/internal/infrastructure/logs/logsEngine"
	"github.com/valyala/fasthttp"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

type Server struct {
	Server    fasthttp.Server
	Stop      chan bool
	IsStarted atomic.Value
	Router    *router.Router
	Config    config.HTTP
	Ctx       context.Context
	Logs      logsEngine.ILogger
	Cc        *appctx.CoreContext
}

func ProvideWebServer(
	ctx *appctx.CoreContext,
	routes *router.Router,
	cfg config.Configuration,
	logs logsEngine.ILogger,
) *Server {
	webServer := new(Server)
	webServer.Router = routes
	webServer.IsStarted.Store(false)
	webServer.Stop = make(chan bool)
	webServer.Config = cfg.Http
	webServer.Ctx = ctx.Ctx()
	webServer.Logs = logs
	webServer.Cc = ctx
	return webServer
}

func (w *Server) ServerStop() {
	if w.IsStarted.Load().(bool) {
		w.Stop <- true
	}
}

func (w *Server) IsServerStarted() bool {
	return w.IsStarted.Load().(bool)
}

func (w *Server) Serve() {
	w.Server = fasthttp.Server{
		Handler:            w.Router.Handler,
		ReadTimeout:        w.Config.GetTimeout(),
		WriteTimeout:       w.Config.GetTimeout(),
		Logger:             w.Logs.Debug(),
		DisableKeepalive:   true,
		TCPKeepalive:       false,
		MaxRequestsPerConn: 1,
		Name:               config.ProjectName,
	}
	go func() {
		w.IsStarted.Store(true)
		err := w.Server.ListenAndServe(":" + w.Config.Port)
		if err != nil {
			if err != http.ErrServerClosed {
				w.Logs.Critical().Fatal("Can't start webserver: " + err.Error())
			} else {
				w.Logs.Trace().Println(err)
			}
		}
	}()

	go func() {
		<-w.Stop
		w.Logs.Trace().Println("WebServer: Stop signal received")
		shutDownCtx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		done := make(chan struct{})
		go func() {
			err := w.Server.Shutdown()
			if err != nil {
				w.Logs.Error().Fatal("WebServer shutdown error: " + err.Error())
			}
			done <- struct{}{}
		}()
		select {
		case <-shutDownCtx.Done():
			w.Logs.Error().Println("WebServer shutdown forced")
		case <-done:
			w.Logs.Trace().Println("WebServer shutdown complete")
		}
		cancel()
		close(w.Stop)
		w.IsStarted.Store(false)
	}()
}

func (w *Server) Run() {
	w.Serve()

	//graceful shutdown
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	stopChannel := make(chan bool)

	go func(ch <-chan os.Signal, st chan<- bool) {
		<-ch
		w.Logs.Trace().Println("STOP received")
		w.ServerStop()
		w.Logs.Trace().Println("WebServer STOP send")
		w.Cc.Cancel()
		w.Logs.Trace().Println("Wait while WebServer stop")
		for w.IsServerStarted() {
			time.Sleep(time.Microsecond)
		}
		w.Logs.Trace().Println("Send STOP")
		st <- true
	}(signalChannel, stopChannel)

	<-stopChannel
}
