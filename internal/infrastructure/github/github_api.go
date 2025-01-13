package github

import (
	"log"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v68/github"
)

type API struct {
	appID      int64
	privateKey []byte
}

func NewAPI(appID int64, privateKey []byte) *API {
	return &API{appID: appID, privateKey: privateKey}
}

func (a *API) NewClient(installationID int64) *github.Client {
	// Shared transport to reuse TCP connections.
	tr := http.DefaultTransport

	// Wrap the shared transport for use with the app ID 1 authenticating with installation ID 99.
	itr, err := ghinstallation.New(tr, a.appID, installationID, []byte(a.privateKey))
	if err != nil {
		log.Fatal(err)
	}

	// Use installation transport with github.com/google/go-github
	client := github.NewClient(&http.Client{Transport: itr})

	return client
}
