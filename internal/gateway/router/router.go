package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/internal/gateway/handlers"
	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/pkg/telemetry"
)

func NewRouter() *mux.Router {
	r := mux.NewRouter()

	r.Use(telemetry.LoggingMiddleware)
	r.Use(telemetry.MetricsMiddleware)

	r.HandleFunc("/generate-playlist", otelhttp.NewHandler(http.HandlerFunc(handlers.GeneratePlaylistHandler), "GeneratePlaylist").ServeHTTP)
	r.HandleFunc("/login", otelhttp.NewHandler(http.HandlerFunc(handlers.LoginHandler), "Login").ServeHTTP)
	r.HandleFunc("/callback", otelhttp.NewHandler(http.HandlerFunc(handlers.CallbackHandler), "Callback").ServeHTTP)
	return r
}
