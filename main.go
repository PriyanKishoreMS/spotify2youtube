package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

type song struct {
	ID        int    `json:"id"`
	Album     string `json:"alubum"`
	Artist    string `json:"artist"`
	TrackName string `json:"TrackName"`
}

var (
	youtubeScopes = []string{youtube.YoutubeScope, youtube.YoutubeUploadScope}
	youtubeConfig *oauth2.Config
)

const (
	clientSecretFile = "client_secret.json"
	tokenFile        = "token.json"
)

func init() {
	b, err := os.ReadFile(clientSecretFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, youtubeScopes...)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	youtubeConfig = config
}

func main() {
	http.HandleFunc("/", handleOAuth2Callback)
	go http.ListenAndServe(":3000", nil)

	authURL := youtubeConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v\n", authURL)

	var authCode string
	fmt.Print("Enter the authorization code: ")
	fmt.Scan(&authCode)
	fmt.Println("herer", authCode)

	token, err := youtubeConfig.Exchange(context.Background(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	spotifytoken, err := config.Token(ctx)
	if err != nil {
		log.Fatalf("couldn't get token: %v", err)
	}

	httpClient := spotifyauth.New().Client(ctx, spotifytoken)
	client := spotify.New(httpClient)
	playlistID := spotify.ID("3jnN5vuhnHAy5ZH81vY7I9")
	playlist, err := client.GetPlaylist(ctx, playlistID)
	if err != nil {
		log.Fatalf("error retrieve playlist data: %v", err)
	}
	playlistName := playlist.Name

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
			createPlaylist(token)

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

	err = os.WriteFile(fmt.Sprintf("%s.json", playlistName), jsonData, 0644)
	if err != nil {
		log.Fatal("error writing to file: ", err)
	}

}

func handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	code := queryParams.Get("code")
	fmt.Fprintf(w, "Authorization code received: %s\n", code)
}

func saveToken(token *oauth2.Token) {
	b, err := json.Marshal(token)
	if err != nil {
		log.Fatalf("Unable to marshal token: %v", err)
	}

	err = os.WriteFile(tokenFile, b, 0600)
	if err != nil {
		log.Fatalf("Unable to save token to file: %v", err)
	}
	fmt.Printf("Token saved to %s\n", tokenFile)
}

func createPlaylist(token *oauth2.Token) {
	service, err := youtube.New(youtubeConfig.Client(context.Background(), token))
	if err != nil {
		log.Fatalf("Error creating YouTube client: %v", err)
	}

	newPlaylist := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       "My New Playlist",
			Description: "A playlist created using the YouTube Data API",
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: "public",
		},
	}

	playlist, err := service.Playlists.Insert([]string{"snippet", "status"}, newPlaylist).Do()
	if err != nil {
		log.Fatalf("Error creating playlist: %v", err)
	}

	fmt.Printf("Playlist created: %s\n", playlist.Id)
}
