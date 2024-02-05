package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

type song struct {
	ID        int    `json:"id"`
	Album     string `json:"alubum"`
	Artist    string `json:"artist"`
	TrackName string `json:"song"`
}

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

	var songs []*song
	for page := 1; ; page++ {
		for _, track := range playlist.Tracks.Tracks {
			count++
			s := song{
				ID:        count,
				Album:     track.Track.Album.Name,
				Artist:    track.Track.Artists[0].Name,
				TrackName: track.Track.Name,
			}

			songs = append(songs, &s)

			log.Printf("Track %d: %s - %s", count, s.TrackName, s.Album)
		}
		err = client.NextPage(ctx, &playlist.Tracks)
		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	jsonData, err := json.MarshalIndent(songs, "", " ")
	if err != nil {
		log.Fatal("error in marshal: ", err)
	}

	err = os.WriteFile("playlist.json", jsonData, 0644)
	if err != nil {
		log.Fatal("error writing to file: ", err)
	}

}
