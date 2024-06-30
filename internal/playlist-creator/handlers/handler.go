package handlers

import (
	"net/http"

	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/internal/playlist-creator/services"
	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/pkg/utils"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("playlist-creator")
	meter  = otel.Meter("playlist-creator")
)

func CreatePlaylistHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "CreatePlaylistHandler")
	defer span.End()

	logrus.WithContext(ctx).Info("Creating playlist")

	response, err := services.CreatePlaylist(r)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Failed to create playlist")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	logrus.WithContext(ctx).Info("Playlist created successfully")
	utils.WriteJSONResponse(w, http.StatusOK, response)
}
