package utils

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

var (
	DBName       string = os.Getenv("DB_DATABASE")
	JWTSecret    string = os.Getenv("JWT_SECRET")
	JWTIssuer    string = os.Getenv("JWT_ISSUER")
	ClientID     string = os.Getenv("GOOGLE_CLIENT_ID")
	ClientSecret string = os.Getenv("GOOGLE_CLIENT_SECRET")
	RedirectURL  string = os.Getenv("GOOGLE_REDIRECT_URL")
)

// var (
// 	OauthConfig = &oauth2.Config{
// 		ClientID:     ClientID,
// 		ClientSecret: ClientSecret,
// 		RedirectURL:  RedirectURL,
// 		Scopes: []string{
// 			"https://www.googleapis.com/auth/userinfo.email",
// 			"https://www.googleapis.com/auth/userinfo.profile",
// 		},
// 		Endpoint: google.Endpoint,
// 	}
// )

var HttpClientConfig = &http.Client{
	Timeout: time.Second * 30,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		ForceAttemptHTTP2:   true,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	},
}

type Config struct {
	Port int
	Env  string
	JWT  struct {
		Secret string
		Issuer string
	}
	RateLimiter struct {
		Rps     int
		Burst   int
		Enabled bool
	}
}
