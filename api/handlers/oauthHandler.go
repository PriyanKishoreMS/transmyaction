package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pascaldekloe/jwt"
	"github.com/priyankishorems/transmyaction/internal/data"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

type AuthResponse struct {
	Username string `json:"username"`
}

var (
	ErrUserUnauthorized = echo.NewHTTPError(http.StatusUnauthorized, "user unauthorized")
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

	accessToken, RefreshToken, err := h.Data.Tokens.GenerateAuthTokens(userinfo.Email, h.Config.JWT.Secret, h.Config.JWT.Issuer)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return err
	}

	user := Cake{
		"email":    userinfo.Email,
		"username": userinfo.Name,
		"avatar":   userinfo.Picture,
	}

	data := Cake{
		"accessToken":  string(accessToken),
		"refreshToken": string(RefreshToken),
		"user":         user,
	}

	tokensJSON, err := json.Marshal(data)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return err
	}

	fmt.Println("tokensJSON here", string(tokensJSON))

	c.SetCookie(&http.Cookie{
		Name:   "tokens",
		Value:  url.QueryEscape(string(tokensJSON)),
		Path:   "/",
		MaxAge: 30 * 24 * 60 * 60, // 1 month
		// Domain:   "localhost",
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	return c.Redirect(http.StatusTemporaryRedirect, "http://localhost:5173/auth-callback")
}

func (h *Handlers) RefreshTokenHandler(c echo.Context) error {

	c.Response().Writer.Header().Add("Vary", "Authorization")

	authorizationHeader := c.Request().Header.Get("Authorization")
	if authorizationHeader == "" {
		err := fmt.Errorf("authorization header not found")
		h.Utils.UserUnAuthorizedResponse(c, err)
		return ErrUserUnauthorized
	}

	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		err := fmt.Errorf("invalid authorization header")
		h.Utils.UserUnAuthorizedResponse(c, err)
		return ErrUserUnauthorized
	}

	token := headerParts[1]

	claims, err := jwt.HMACCheck([]byte(token), []byte(h.Config.JWT.Secret))
	if err != nil {
		h.Utils.UserUnAuthorizedResponse(c, err)
		return ErrUserUnauthorized
	}

	id := claims.Subject

	accessToken, err := data.GenerateAccessToken(id, []byte(h.Config.JWT.Secret), h.Config.JWT.Issuer)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return err
	}
	return c.JSON(200, Cake{"accessToken": string(accessToken)})
}
