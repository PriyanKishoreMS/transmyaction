package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func GmailService(token *oauth2.Token) (*gmail.Service, error) {
	client := oauthConfig.Client(context.Background(), token)
	srv, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func (h *Handlers) updateTokens(token *oauth2.Token, email string) (*oauth2.Token, error) {

	ts := oauthConfig.TokenSource(context.Background(), token)

	refreshedToken, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	if refreshedToken.AccessToken != token.AccessToken || refreshedToken.Expiry != token.Expiry {
		err := h.Data.Tokens.UpdateTokens(
			refreshedToken,
			email,
		)
		if err != nil {
			return nil, err
		}
	}

	return refreshedToken, nil
}

func (h *Handlers) GmailHandler(c echo.Context) error {

	email := c.Param("email")

	token, err := h.Data.Tokens.GetTokenFromEmail(email)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("can't get token from db: %v", err)
	}

	refreshedToken, err := h.updateTokens(token, email)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("can't refresh token: %v", err)
	}

	token = refreshedToken

	srv, err := GmailService(token)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("can't get gmail service: %v", err)
	}

	user := "me"
	rLabels, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("unable to get label: %v", err)
	}

	fmt.Println("Labels:")
	for _, l := range rLabels.Labels {
		fmt.Println(l.Name)
	}

	return c.JSON(http.StatusOK, Cake{
		"message": "wokring",
	})
}
