package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/pkg/auth"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("gateway")
	meter  = otel.Meter("gateway")
)

func GeneratePlaylist(r *http.Request) (map[string]interface{}, error) {
	ctx, span := tracer.Start(r.Context(), "GeneratePlaylistService")
	defer span.End()

	_, authSpan := tracer.Start(ctx, "Authenticator")

	client, err := auth.GetClient()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Client not authenticated")
		return nil, errors.New("client not authenticated")
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Error reading request body")
		return nil, err
	}
	defer r.Body.Close()

	token, err := client.Token()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Couldn't get token")
		return nil, errors.New("couldn't get token")
	}

	tokenBytes, err := json.Marshal(token)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Couldn't marshal token")
		return nil, errors.New("couldn't marshal token")
	}

	authSpan.End()

	reqBody := map[string]interface{}{
		"taste": json.RawMessage(bodyBytes),
		"token": json.RawMessage(tokenBytes),
	}
	jsonReqBody, err := json.Marshal(reqBody)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Couldn't marshal request body")
		return nil, errors.New("couldn't marshal request body")
	}

	_, playlistSpan := tracer.Start(ctx, "RequestPlaylistCreation")

	httpClient := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	req, err := http.NewRequestWithContext(ctx, "POST", "http://playlist-creator:8082/create-playlist", bytes.NewReader(jsonReqBody))
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Error creating request to playlist creation service")
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Error sending request to playlist creation service")
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Error reading response from playlist creation service")
		return nil, err
	}

	playlistSpan.End()

	var response map[string]interface{}
	if err := json.Unmarshal(respBody, &response); err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Error unmarshalling response from playlist creation service")
		return nil, err
	}

	logrus.WithContext(ctx).Info("Playlist generated successfully")
	return response, nil
}
