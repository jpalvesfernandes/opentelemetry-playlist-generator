package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/internal/playlist-creator/router"
	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/pkg/telemetry"
)

func main() {
	port := "8082"
	ctx := context.Background()
	otelShutdown, err := telemetry.InitOtel(ctx, "playlist-creator")
	if err != nil {
		return
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	r := router.NewRouter()

	if err := http.ListenAndServe(":"+port, r); err != nil {
		fmt.Printf("Failed to start server: %v", err)
	}
}
