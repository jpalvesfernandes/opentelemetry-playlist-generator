package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/jpalvesfernandes/opentelemetry-playlist-generator/internal/song-matcher/models"
	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	tracer = otel.Tracer("song-matcher")
	meter  = otel.Meter("song-matcher")
	client *spotify.Client
)

func InitSpotifyClient() {
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(context.Background())
	if err != nil {
		fmt.Printf("couldn't get token: %v\n", err)
	}

	httpClient := spotifyauth.New().Client(context.Background(), token)
	client = spotify.New(httpClient)
}

func MatchSongs(r *http.Request) ([]models.Song, error) {
	ctx, span := tracer.Start(r.Context(), "MatchSongsService")
	defer span.End()

	var taste models.Taste

	err := json.NewDecoder(r.Body).Decode(&taste)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("Error decoding request body")
		return nil, errors.New("error decoding request body")
	}
	defer r.Body.Close()

	logrus.WithContext(ctx).WithFields(logrus.Fields{
		"Artists": taste.FavoriteArtists,
		"Genres":  taste.Genres,
	}).Info("Matching songs based on taste")

	favoriteArtistCount, _ := meter.Int64Counter("favorite.artists.total", metric.WithDescription("Total number of favorite artists"))
	favoriteGenreCount, _ := meter.Int64Counter("favorite.genres.total", metric.WithDescription("Total number of favorite genres"))

	for _, a := range taste.FavoriteArtists {
		attrs := []attribute.KeyValue{
			attribute.String("artist", a),
		}
		favoriteArtistCount.Add(ctx, 1, metric.WithAttributes(attrs...))
	}
	for _, t := range taste.Genres {
		attrs := []attribute.KeyValue{
			attribute.String("genre", t),
		}
		favoriteGenreCount.Add(ctx, 1, metric.WithAttributes(attrs...))
	}

	_, artistSpan := tracer.Start(ctx, "SearchFavoriteArtists")

	var artistIDs []spotify.ID
	for _, artist := range taste.FavoriteArtists {
		artists, err := client.Search(context.Background(), artist, spotify.SearchTypeArtist)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).WithFields(logrus.Fields{
				"Artists": taste.FavoriteArtists,
				"Genres":  taste.Genres,
			}).Info("Invalid taste data: favorite artists or genre missing")
			return nil, fmt.Errorf("error searching artist: %v", err)
		}
		if len(artists.Artists.Artists) > 0 {
			artistIDs = append(artistIDs, artists.Artists.Artists[0].ID)
		}
	}

	logrus.WithContext(ctx).WithFields(logrus.Fields{
		"Artists": taste.FavoriteArtists,
		"Genres":  taste.Genres,
	}).Info("Artists and Genre matching taste found")

	artistSpan.End()

	_, recommendationSpan := tracer.Start(ctx, "GetRecommendations")

	seeds := spotify.Seeds{
		Artists: artistIDs,
		Genres:  taste.Genres,
	}

	trackAttributes := spotify.NewTrackAttributes().
		MaxValence(0.4).
		TargetEnergy(0.6).
		TargetDanceability(0.6)

	res, err := client.GetRecommendations(context.Background(), seeds, trackAttributes, spotify.Country("US"), spotify.Limit(30))
	if err != nil {
		logrus.WithContext(ctx).WithError(err).WithFields(logrus.Fields{
			"Artists": taste.FavoriteArtists,
			"Genres":  taste.Genres,
		}).Error("Error finding songs matching taste")
		return nil, fmt.Errorf("error getting recommendations: %v", err)
	}

	var songs []models.Song
	for _, track := range res.Tracks {
		songs = append(songs, models.Song{
			Title:  track.Name,
			Artist: track.Artists[0].Name,
			ID:     string(track.ID),
		})
	}

	recommendationSpan.End()

	logrus.WithContext(ctx).WithFields(logrus.Fields{
		"Artists": taste.FavoriteArtists,
		"Genres":  taste.Genres,
	}).Info("Songs recommendation matched successfully")

	return songs, nil
}
