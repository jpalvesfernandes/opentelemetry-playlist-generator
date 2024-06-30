package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/internal/playlist-creator/models"
	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/pkg/auth"
	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"golang.org/x/oauth2"
)

var (
	tracer = otel.Tracer("playlist-creator")
	meter  = otel.Meter("playlist-creator")
)

func CreatePlaylist(r *http.Request) (map[string]string, error) {
	ctx, span := tracer.Start(r.Context(), "CreatePlaylistService")
	defer span.End()

	var playlistRequest models.Request

	err := json.NewDecoder(r.Body).Decode(&playlistRequest)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Error decoding request body")
		return nil, errors.New("error decoding request body")
	}
	defer r.Body.Close()

	var token oauth2.Token
	err = json.Unmarshal(playlistRequest.Token, &token)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Invalid token")
		return nil, errors.New("invalid token")
	}

	var taste models.Taste
	err = json.Unmarshal(playlistRequest.Taste, &taste)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"Artists": taste.FavoriteArtists,
			"Genres":  taste.Genres,
		}).WithError(err).Error("Invalid user data")
		return nil, errors.New("invalid user data")
	}

	_, authSpan := tracer.Start(ctx, "Authenticator")

	authenticator := auth.GetAuthenticator()
	ac := authenticator.Client(context.Background(), &token)
	sc := spotify.New(ac)

	authSpan.End()

	songMatcherReqBody, err := json.Marshal(map[string]interface{}{
		"favorite_artists": taste.FavoriteArtists,
		"genres":           taste.Genres,
	})
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Couldn't marshal request body for song matcher")
		return nil, errors.New("couldn't marshal request body for song matcher")
	}

	_, songMatcherSpan := tracer.Start(ctx, "SongMatcher")

	httpClient := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	req, err := http.NewRequestWithContext(ctx, "POST", "http://song-matcher:8081/match-songs", bytes.NewReader(songMatcherReqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error requesting songs from song matcher: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"status_code": resp.StatusCode}).WithError(err).Error("Error making request to song matcher")
		return nil, fmt.Errorf("error reading response from song matcher: %v", err)
	}

	var songs []models.Song
	err = json.Unmarshal(respBody, &songs)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Error decoding song matcher response")
		return nil, errors.New("invalid songs data from song matcher")
	}

	songMatcherSpan.End()

	_, playlistSpan := tracer.Start(ctx, "CreatePlaylist")

	user, err := sc.CurrentUser(context.Background())
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Failed access Spotify user data")
		return nil, err
	}

	playlist, err := sc.CreatePlaylistForUser(context.Background(), user.ID, "Recommended Songs", "Playlist created from recommended songs", false, false)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Failed to create playlist on Spotify to user")
		return nil, err
	}

	playlistSpan.End()

	_, trackSpan := tracer.Start(ctx, "AddTracksToPlaylist")

	var trackIDs []spotify.ID
	for _, song := range songs {
		trackIDs = append(trackIDs, spotify.ID(song.ID))
	}

	_, err = sc.AddTracksToPlaylist(context.Background(), playlist.ID, trackIDs...)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"playlist_id": playlist.ID,
			"tracks_id":   trackIDs,
		}).WithError(err).Error("Failed to add tracks to playlist on Spotify")
		return nil, err
	}

	trackSpan.End()

	playlistCreated, _ := meter.Int64Counter("playlists.created.total", metric.WithDescription("Total number of created playlists"))
	playlistCreated.Add(ctx, 1)

	logrus.WithContext(ctx).WithFields(logrus.Fields{
		"playlist_id": playlist.ID,
		"tracks_id":   trackIDs,
	}).Info("Playlist created successfully with tracks")

	return map[string]string{"message": "Playlist created successfully", "playlist_id": string(playlist.ID)}, nil
}
