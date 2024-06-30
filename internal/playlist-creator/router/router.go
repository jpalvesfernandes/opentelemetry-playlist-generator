package router

import (
	"net/http"

	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/internal/playlist-creator/handlers"
	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/pkg/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	r := mux.NewRouter()

	r.Use(telemetry.LoggingMiddleware)
	r.Use(telemetry.MetricsMiddleware)

	r.HandleFunc("/create-playlist", otelhttp.NewHandler(http.HandlerFunc(handlers.CreatePlaylistHandler), "CreatePlaylist").ServeHTTP)

	return r
}
