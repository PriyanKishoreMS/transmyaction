package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var oauthConfig *oauth2.Config

func (h *Handlers) InitOAuth() {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("unable to read client secret: %v", err)
	}

	config, err := google.ConfigFromJSON(b, "openid", "email", "profile", gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("unable to parse client secret file: %v", err)
	}

	oauthConfig = config
}

func (h *Handlers) LoginHandler(c echo.Context) error {
	url := oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

type GoogleUserInfo struct {
	Email         string `json:"email"`
	Name          string `json:"name"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
}

func getUserEmail(token *oauth2.Token) (*GoogleUserInfo, error) {
	client := oauthConfig.Client(context.Background(), token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get userinfo: %v", err)
	}
	defer resp.Body.Close()

	fmt.Println("Resp: ", resp)

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo: %v", err)
	}

	return &userInfo, nil
}

func (h *Handlers) CallbackHandler(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		h.Utils.BadRequest(c, fmt.Errorf("code not found in query params"))
		return fmt.Errorf("code not found in query params")
	}

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("unable to exchange token: " + err.Error())
	}

	fmt.Println("token: ", token)

	userinfo, err := getUserEmail(token)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("failed to fetch email: " + err.Error())
	}

	fmt.Println("User email:", userinfo)

	fmt.Println("refresh_tokem: ", token.RefreshToken)

	err = h.Data.Tokens.SaveEmailToken(userinfo.Email, userinfo.Name, token)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("can't' save token: %v", err)
	}

	return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173")
}

// func saveToken(token *oauth2.Token, email string) error {
// 	filename := "tokens.json"

// 	file, err := os.Create(filename)
// 	if err != nil {
// 		return fmt.Errorf("can't create file: %v", err)
// 	}
// 	defer file.Close()

// 	json.NewEncoder(file).Encode(token)

// 	return nil
// }

// client := oauthConfig.Client(context.Background(), tok)
// srv, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
// if err != nil {
// 	return c.String(http.StatusInternalServerError, "unable to create Gmail client: "+err.Error())
// }

// user := "me"
// msgs, err := srv.Users.Messages.List(user).
// 	MaxResults(10).
// 	Do()
// if err != nil {
// 	return c.String(http.StatusInternalServerError, "unable to list messages: "+err.Error())
// }
// var res string

// for _, m := range msgs.Messages {
// 	fullMsg, err := srv.Users.Messages.Get(user, m.Id).Do()
// 	if err != nil {
// 		return c.String(http.StatusInternalServerError, "unable to get message: "+err.Error())
// 	}

// 	var from, subject string
// 	for _, h := range fullMsg.Payload.Headers {
// 		switch h.Name {
// 		case "From":
// 			from = h.Value
// 		case "Subject":
// 			subject = h.Value
// 		}
// 	}

// 	var body string
// 	if fullMsg.Payload.Body != nil && fullMsg.Payload.Body.Data != "" {
// 		decoded, _ := base64.URLEncoding.DecodeString(fullMsg.Payload.Body.Data)
// 		body = string(decoded)
// 	} else if len(fullMsg.Payload.Parts) > 0 {
// 		for _, part := range fullMsg.Payload.Parts {
// 			if part.MimeType == "text/plain" {
// 				decoded, _ := base64.URLEncoding.DecodeString(part.Body.Data)
// 				body = string(decoded)
// 				break
// 			}
// 		}
// 	}

// 	res += fmt.Sprintf("From: %s\nSubject: %s\nBody: %s\n\n", from, subject, body)
// }
