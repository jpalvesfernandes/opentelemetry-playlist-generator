package router

import (
	"net/http"

	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/internal/song-matcher/handlers"
	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/pkg/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	r := mux.NewRouter()

	r.Use(telemetry.LoggingMiddleware)
	r.Use(telemetry.MetricsMiddleware)

	r.HandleFunc("/match-songs", otelhttp.NewHandler(http.HandlerFunc(handlers.MatchSongsHandler), "MatchSongs").ServeHTTP)

	return r
}
