package routes

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/satmaelstorm/filup/internal/infrastructure/logs/logsEngine"
	"net/http"
	"runtime/debug"
)

func ProvideRoutes(logger logsEngine.Loggers) *mux.Router {
	r := mux.NewRouter()
	s := r.PathPrefix("/").Subrouter()

	s.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					logger.Critical().Printf("%s : %s\n", err, string(debug.Stack()))
				}
			}()
			next.ServeHTTP(w, req)
		})
	})

	s.HandleFunc(Upload, func(writer http.ResponseWriter, request *http.Request) {
		_ = request.ParseMultipartForm(1024 * 1024 * 100)
		_ = request.Form
		defer request.Body.Close()
	}).Methods(methodPOST)
	s.HandleFunc(Metrics, promhttp.Handler().ServeHTTP).Methods(methodGET)

	return r
}
