package api

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/luizbafilho/fusis/types"
)

type ErrResponse map[string]interface{}

func CustomHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	response := ErrResponse{}

	switch v := err.(type) {
	case types.ErrNotFound:
		code = http.StatusNotFound
		response = ErrResponse{"error": v.Error()}
	case types.ErrConflict:
		code = http.StatusConflict
		response = ErrResponse{"error": v.Error()}
	case types.ErrValidation:
		code = http.StatusUnprocessableEntity
		response = ErrResponse{"error": v.Errors}
	case *echo.HTTPError:
		response = ErrResponse{"error": v.Message}
	default:
		response = ErrResponse{"error": v.Error()}
	}

	if !c.Response().Committed {
		if c.Request().Method == echo.HEAD { // Issue #608
			c.NoContent(code)
		} else {
			c.JSON(code, response)
		}
	}
}
