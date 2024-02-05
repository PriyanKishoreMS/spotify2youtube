package main

import (
	"context"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

func main() {
	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		log.Fatalf("couldn't get token: %v", err)
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	playlistID := spotify.ID("3jnN5vuhnHAy5ZH81vY7I9")
	playlist, err := client.GetPlaylist(ctx, playlistID)
	if err != nil {
		log.Fatalf("error retrieve playlist data: %v", err)
	}

	count := 0
	log.Printf("Playlist %s has a total of %d tracks", playlist.Name, playlist.Tracks.Total)
	for page := 1; ; page++ {
		for id, track := range playlist.Tracks.Tracks {
			count++
			log.Printf("Track %d: %s - %s", id, track.Track.Name, track.Track.Album.Name)
		}
		err = client.NextPage(ctx, &playlist.Tracks)
		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	}

}
