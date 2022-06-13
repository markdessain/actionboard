package github

import (
	"actionboard/app/data"
	httpl "actionboard/app/http"
	"context"
	"github.com/google/go-github/v44/github"
	"golang.org/x/oauth2"
	"net/http"
)

// GetClient creates a new authenticated GitHub client which will cache HTTP responses
// and use the etags to reduce the amount of API calls that need to be made.
func GetClient(d *data.Data, token string) *github.Client {

	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	oauthTransport := oauth2.NewClient(ctx, ts).Transport
	mainTransport := httpl.NewTransport(&d.HttpCache)
	mainTransport.Transport = oauthTransport

	tx := &http.Client{
		Transport: mainTransport,
	}

	return github.NewClient(tx)

}