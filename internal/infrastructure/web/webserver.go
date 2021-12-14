package web

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/satmaelstorm/filup/internal/infrastructure/config"
	"github.com/satmaelstorm/filup/internal/infrastructure/ctx"
	"github.com/satmaelstorm/filup/internal/infrastructure/logs/logsEngine"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

type Server struct {
	Server    http.Server
	Stop      chan bool
	IsStarted atomic.Value
	Router    *mux.Router
	Config    config.HTTP
	Ctx       context.Context
	Logs      logsEngine.Loggers
}

func ProvideWebServer(
	ctx *ctx.CoreContext,
	routes *mux.Router,
	cfg config.Configuration,
	logs logsEngine.Loggers,
) *Server {
	webServer := new(Server)
	webServer.Router = routes
	webServer.IsStarted.Store(false)
	webServer.Stop = make(chan bool)
	webServer.Config = cfg.Http
	webServer.Ctx = ctx.Ctx()
	webServer.Logs = logs
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
	w.Server = http.Server{
		Addr:              ":" + w.Config.Port,
		Handler:           w.Router,
		ReadTimeout:       w.Config.GetTimeout(),
		ReadHeaderTimeout: w.Config.GetTimeout(),
		WriteTimeout:      w.Config.GetTimeout(),
		ErrorLog:          w.Logs.G(logsEngine.LogError),
	}
	go func() {
		w.IsStarted.Store(true)
		err := w.Server.ListenAndServe()
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
			err := w.Server.Shutdown(shutDownCtx)
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
		w.Logs.Trace().Println("STOP")
		w.ServerStop()
		w.Logs.Trace().Println("Http - сигнал СТОП отправлен")

		w.Logs.Trace().Println("Ждем остановки http сервера")
		for w.IsServerStarted() {
			time.Sleep(time.Microsecond)
		}
		w.Logs.Trace().Println("Отправка общего сигнала СТОП")
		st <- true
	}(signalChannel, stopChannel)

	<-stopChannel
}
