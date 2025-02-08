package lib

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Client struct {
	httpClient *http.Client
	projectID  string
	location   string
}

func NewClient(httpClient *http.Client, projectID, location string) *Client {
	return &Client{
		httpClient: httpClient,
		projectID:  projectID,
		location:   location,
	}
}

func NewClientWithCredentials(credentials *google.Credentials, projectID, location string) *Client {
	ctx := context.Background()
	client := oauth2.NewClient(ctx, credentials.TokenSource)

	return NewClient(client, projectID, location)
}
