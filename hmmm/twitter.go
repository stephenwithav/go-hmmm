package main

// OAuth1
import (
	"net/http"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// createTwitterOathClient ...
func createTwitterOathClient(consumerKey, consumerSecret, accessToken, accessSecret string) *twitter.Client {
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	// http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	return twitter.NewClient(httpClient)
}

// sendTweet ...
func sendTweet(client *twitter.Client, status string, replyingTo int64) (*twitter.Tweet, *http.Response, error) {
	if replyingTo > 0 {
		return client.Statuses.Update(status, &twitter.StatusUpdateParams{InReplyToStatusID: replyingTo})
	}

	return client.Statuses.Update(status, nil)
}
