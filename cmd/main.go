package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/gommon/log"
	"github.com/priyankishorems/transmyaction/api"
	"github.com/priyankishorems/transmyaction/api/handlers"
	"github.com/priyankishorems/transmyaction/utils"
)

var validate validator.Validate

func main() {
	cfg := &utils.Config{}

	flag.IntVar(&cfg.Port, "port", 3000, "Server port")
	flag.StringVar(&cfg.Env, "env", "development", "Server port")

	flag.IntVar(&cfg.RateLimiter.Rps, "limiter-rps", 50, "Rate limiter max requests per second")
	flag.IntVar(&cfg.RateLimiter.Burst, "limiter-burst", 50, "Rate limiter max burst")
	flag.StringVar(&cfg.JWT.Secret, "jwt-secret", utils.JWTSecret, "JWT secret")
	flag.StringVar(&cfg.JWT.Issuer, "jwt-issuer", utils.JWTIssuer, "JWT issuer")
	flag.BoolVar(&cfg.RateLimiter.Enabled, "limiter-enabled", false, "Rate limiter enabled")

	flag.Parse()
	log.SetHeader("${time_rfc3339} ${level}")

	// db := data.PSQLDB{}
	// dbPool, err := db.Open()
	// if err != nil {
	// 	log.Fatalf("error in opening db; %v", err)
	// }
	// defer dbPool.Close()

	validate = *validator.New()

	h := &handlers.Handlers{
		Config:   *cfg,
		Validate: validate,
		Utils:    utils.NewUtils(),
		// Data:     data.NewModel(dbPool),
		// RedditBot: redditBot,
	}

	e := api.SetupRoutes(h)
	e.Server.ReadTimeout = time.Second * 10
	e.Server.WriteTimeout = time.Second * 20
	e.Server.IdleTimeout = time.Minute
	e.HideBanner = true
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", cfg.Port)))
}
