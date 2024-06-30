package auth

import (
	"errors"
	"net/http"
	"sync"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var (
	authenticator = spotifyauth.New(
		spotifyauth.WithRedirectURL("http://localhost:8080/callback"),
		spotifyauth.WithScopes(
			spotifyauth.ScopePlaylistModifyPublic,
			spotifyauth.ScopePlaylistModifyPrivate,
		),
	)
	client      *spotify.Client
	clientMutex sync.Mutex
)

func GetAuthURL(state string) string {
	return authenticator.AuthURL(state)
}

func CompleteAuth(r *http.Request, state string) error {
	tok, err := authenticator.Token(r.Context(), state, r)
	if err != nil {
		return err
	}

	httpClient := authenticator.Client(r.Context(), tok)
	spotifyClient := spotify.New(httpClient)

	clientMutex.Lock()
	defer clientMutex.Unlock()
	client = spotifyClient

	return nil
}

func GetClient() (*spotify.Client, error) {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	if client == nil {
		return nil, errors.New("client not authenticated")
	}
	return client, nil
}

func GetAuthenticator() *spotifyauth.Authenticator {
	return authenticator
}
