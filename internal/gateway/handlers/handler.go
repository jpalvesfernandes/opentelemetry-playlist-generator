package handlers

import (
	"net/http"

	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/internal/gateway/services"
	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/pkg/auth"
	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/pkg/utils"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("gateway")
	meter  = otel.Meter("gateway")
	state  = "abc123"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "LoginHandler")
	defer span.End()

	url := auth.GetAuthURL(state)
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"url": url})
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "CallbackHandler")
	defer span.End()

	if err := auth.CompleteAuth(r, state); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Failed to complete auth")
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "Couldn't complete auth")
		return
	}

	logrus.WithContext(ctx).Info("Auth completed successfully")
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Login Completed!"})
}

func GeneratePlaylistHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "GeneratePlaylistHandler")
	defer span.End()

	response, err := services.GeneratePlaylist(r)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Failed to generate playlist")
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	logrus.WithContext(ctx).Info("Playlist generated successfully")
	utils.WriteJSONResponse(w, http.StatusOK, response)
}
