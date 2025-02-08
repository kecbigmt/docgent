package rag

import "golang.org/x/oauth2/google"

type Config struct {
	ProjectID   string
	Location    string
	Credentials *google.Credentials
}

func NewConfig(projectID, location string, credentials *google.Credentials) Config {
	return Config{
		ProjectID:   projectID,
		Location:    location,
		Credentials: credentials,
	}
}
