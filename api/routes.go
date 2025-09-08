package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/priyankishorems/transmyaction/api/handlers"
)

func SetupRoutes(h *handlers.Handlers) *echo.Echo {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		// AllowOrigins:     []string{"http://localhost:5173"},
		AllowCredentials: true,
	}))
	e.Use(IPRateLimit(h))
	e.Use(middleware.RemoveTrailingSlash())
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		c.Logger().Error(err)
		e.DefaultHTTPErrorHandler(err, c)
	}

	e.HideBanner = true

	h.InitOAuth()
	e.GET("/", h.HomeFunc)
	e.GET("/login", h.LoginHandler)
	e.GET("/oauth2/callback", h.CallbackHandler)
	e.GET("/refresh", h.RefreshTokenHandler, Authenticate(*h))

	e.POST("/txns/:email", h.UpdateTransactionsHandler, Authenticate(*h))
	e.POST("/txns/add", h.InsertTransactionHandler, Authenticate(*h))
	e.GET("/txns/:email/:interval/:year/:month", h.GetTransactionsHandler, Authenticate(*h))
	e.GET("/txns/:email", h.GetTransactionsHandler, Authenticate(*h))
	e.POST("/txns/update", h.UpdateTransactionsHandler, Authenticate(*h))

	// api := e.Group("/api")
	// {

	// }

	// istLocation, err := time.LoadLocation("Asia/Kolkata")
	// if err != nil {
	// 	log.Fatal("Error loading IST timezone:", err)
	// }

	// scheduler, err := gocron.NewScheduler(
	// 	gocron.WithLocation(istLocation),
	// )
	// if err != nil {
	// 	log.Fatal("Error creating scheduler", err)
	// }

	// updateTxnsAtTime := gocron.NewAtTime(00, 00, 00)
	// updateTxnsAtTimes := gocron.NewAtTimes(updateTxnsAtTime)

	// updateTxnsnJob, err := jobs.UpdateTxnsJob(*h, scheduler, updateTxnsAtTimes)
	// if err != nil {
	// 	log.Fatal("Error creating job: ", err)
	// }

	// fmt.Println("updateTransactionsJob started: ", updateTxnsnJob.ID())

	// scheduler.Start()

	return e
}
