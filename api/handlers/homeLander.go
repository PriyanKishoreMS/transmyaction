package handlers

import (
	"fmt"
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
	// msg := Cake{
	// 	"message": "Welcome to Transmyaction API",
	// 	"status":  "available",
	// 	"system_info": Cake{
	// 		"environment": h.Config.Env,
	// 		"port":        h.Config.Port,
	// 	},
	// }
	// return c.JSON(http.StatusOK, msg)

	distinctEmails, err := h.Data.Txns.GetAllDistinctEmails()
	if err != nil {
		return nil
	}

	for _, email := range distinctEmails {
		fmt.Println("Distinct email found:", FormatTimeForAxis(email.LastUpdated))
	}

	return c.JSON(http.StatusOK, distinctEmails)
}
