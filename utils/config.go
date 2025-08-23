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
	DBName     string = os.Getenv("DB_DATABASE")
	DBUsername string = os.Getenv("DB_USERNAME")
	DBPassword string = os.Getenv("DB_PASSWORD")
	DBPort     string = os.Getenv("DB_PORT")
	DBHost     string = os.Getenv("DB_HOST")
	JWTSecret  string = os.Getenv("JWT_SECRET")
	JWTIssuer  string = os.Getenv("JWT_ISSUER")
)

// var (
// 	OauthConfig = &oauth2.Config{
// 		ClientID:     RedditIdWeb,
// 		ClientSecret: RedditSecretWeb,
// 		RedirectURL:  RedirectURL,
// 		Endpoint: oauth2.Endpoint{
// 			AuthURL:  "https://www.reddit.com/api/v1/authorize",
// 			TokenURL: "https://www.reddit.com/api/v1/access_token",
// 		},
// 		Scopes: []string{"identity", "read"},
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
