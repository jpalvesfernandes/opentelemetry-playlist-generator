package handlers

import (
	"net/http"

	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/internal/song-matcher/services"
	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/pkg/utils"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("song-matcher")
	meter  = otel.Meter("song-matcher")
)

func MatchSongsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "MatchSongsHandler")
	defer span.End()

	response, err := services.MatchSongs(r)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Failed to match songs")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	logrus.WithContext(ctx).Info("Songs matched successfully")
	utils.WriteJSONResponse(w, http.StatusOK, response)
}
