package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/nstehr/pitwall/cli/internal/config"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/oauth2"
)

const (
	clientID = "pitwall-cli"
)

func Authenticate(ctx context.Context, authEndpoint string) (*oauth2.Token, error) {
	// based on: https://gist.github.com/marians/3b55318106df0e4e648158f1ffb43d38
	serverStopChan := make(chan struct{})
	conf := &oauth2.Config{
		ClientID: clientID,
		Scopes:   []string{"openid", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/realms/pitwall/protocol/openid-connect/auth", authEndpoint),
			TokenURL: fmt.Sprintf("%s/realms/pitwall/protocol/openid-connect/token", authEndpoint),
		},
		// my own callback URL
		RedirectURL: "http://127.0.0.1:9898/oauth/callback",
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	sslcli := &http.Client{Transport: tr}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, sslcli)

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	authCodeUrl := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)

	log.Println("You will now be taken to your browser for authentication")
	time.Sleep(1 * time.Second)
	open.Run(authCodeUrl)
	time.Sleep(1 * time.Second)

	var token *oauth2.Token
	var handleError error
	// the anonymous handler is to capture the variables we've define that we need
	// to be able to return the tokens as a method call
	http.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		queryParts, _ := url.ParseQuery(r.URL.RawQuery)

		// Use the authorization code that is pushed to the redirect
		// URL.
		code := queryParts["code"][0]

		// Exchange will do the handshake to retrieve the initial access token.
		tok, err := conf.Exchange(ctx, code)
		if err != nil {
			handleError = err
		}
		token = tok
		// show succes page
		msg := "<p><strong>Success!</strong></p>"
		msg = msg + "<p>You are authenticated and can now return to the CLI.</p>"
		fmt.Fprint(w, msg)
		serverStopChan <- struct{}{}
	})

	go http.ListenAndServe(":9898", nil)
	// TODO: should also have a timeout
	<-serverStopChan
	return token, handleError
}

func getAuthenticatedClient(ctx context.Context, cfg *config.Config) *http.Client {
	conf := &oauth2.Config{
		ClientID: clientID,
		Scopes:   []string{"openid", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/realms/pitwall/protocol/openid-connect/auth", cfg.AuthEndpoint),
			TokenURL: fmt.Sprintf("%s/realms/pitwall/protocol/openid-connect/token", cfg.AuthEndpoint),
		},
	}
	token := &oauth2.Token{}
	token.AccessToken = cfg.OAuth.AccessToken
	token.RefreshToken = cfg.OAuth.RefreshToken
	token.Expiry = cfg.OAuth.Expiry
	token.TokenType = cfg.OAuth.TokenType
	tokenSource := conf.TokenSource(ctx, token)
	client := oauth2.NewClient(ctx, tokenSource)
	return client
}
