package models

type Taste struct {
	FavoriteArtists []string `json:"favorite_artists"`
	Genres          []string `json:"genres"`
}