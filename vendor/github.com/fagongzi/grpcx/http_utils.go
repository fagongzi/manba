package grpcx

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"
)

// JSONResult json result
type JSONResult struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

// NewJSONBodyHTTPHandle returns a http handle JSON body
func NewJSONBodyHTTPHandle(factory func() interface{}, handler func(interface{}) (*JSONResult, error)) func(echo.Context) error {
	return func(ctx echo.Context) error {
		value := factory()
		err := ReadJSONFromBody(ctx, value)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, &JSONResult{
				Data: err.Error(),
			})
		}

		result, err := handler(value)
		if err != nil {
			return ctx.NoContent(http.StatusInternalServerError)
		}

		return ctx.JSON(http.StatusOK, result)
	}
}

// NewGetHTTPHandle return get http handle
func NewGetHTTPHandle(factory func(echo.Context) (interface{}, error), handler func(interface{}) (*JSONResult, error)) func(echo.Context) error {
	return func(ctx echo.Context) error {
		value, err := factory(ctx)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, &JSONResult{
				Data: err.Error(),
			})
		}

		result, err := handler(value)
		if err != nil {
			return ctx.NoContent(http.StatusInternalServerError)
		}

		return ctx.JSON(http.StatusOK, result)
	}
}

// ReadJSONFromBody read json body
func ReadJSONFromBody(ctx echo.Context, value interface{}) error {
	data, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}

	if len(data) > 0 {
		err = json.Unmarshal(data, value)
		if err != nil {
			return err
		}
	}

	return nil
}
