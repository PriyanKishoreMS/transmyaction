package handlers

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/priyankishorems/transmyaction/internal/data"
	"github.com/priyankishorems/transmyaction/utils"
)

type Cake map[string]interface{}
type Handlers struct {
	Config   utils.Config
	Validate validator.Validate
	Utils    utils.Utilities
	Data     data.Models
}

func (h *Handlers) HomeFunc(c echo.Context) error {
	msg := Cake{
		"message": "Welcome to Transmyaction API",
		"status":  "available",
		"system_info": Cake{
			"environment": h.Config.Env,
			"port":        h.Config.Port,
		},
	}
	return c.JSON(http.StatusOK, msg)
}
