package resp

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type JSONSerializer struct{}

// Serialize converts an interface into a json and writes it to the response.
// You can optionally use the indent parameter to produce pretty JSONs.
func (d *JSONSerializer) Serialize(c echo.Context, i interface{}, indent string) error {
	enc := json.NewEncoder(c.Response())
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(i)
}

// Deserialize reads a JSON from a request body and converts it into an interface.
func (d *JSONSerializer) Deserialize(c echo.Context, i interface{}) error {
	err := json.NewDecoder(c.Request().Body).Decode(i)
	return err
}

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func PageOK(c echo.Context, data interface{}, total int64) error {
	return c.JSON(http.StatusOK, &Result{Code: 200, Data: H{"total": total, "list": data}})
}

func Ok(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, &Result{Code: 200, Data: data})
}

func ParamErr(c echo.Context, msg string, a ...any) error {
	return c.JSON(http.StatusOK, &Result{Code: 400, Msg: "参数解析错误:" + fmt.Sprintf(msg, a...)})
}
func BadRequest(c echo.Context, msg string, a ...any) error {
	return c.JSON(http.StatusOK, &Result{Code: 400, Msg: fmt.Sprintf(msg, a...)})
}
func NotFound(c echo.Context, msg string, a ...any) error {
	return c.JSON(http.StatusOK, &Result{Code: 404, Msg: fmt.Sprintf(msg, a...)})
}

func NoPermission(c echo.Context) error {
	return c.JSON(http.StatusOK, &Result{Code: 403})
}

func NoLogin(c echo.Context) error {
	return c.JSON(http.StatusUnauthorized, &Result{Code: 401})
}

func ServerErr(c echo.Context, msg string, a ...any) error {
	return c.JSON(http.StatusOK, &Result{Code: 500, Msg: fmt.Sprintf(msg, a...)})
}
func Json(c echo.Context, code int, data interface{}, msg string, a ...any) error {
	return c.JSON(http.StatusOK, &Result{Code: code, Data: data, Msg: fmt.Sprintf(msg, a...)})
}
func Resp(c echo.Context, code int, data interface{}) error {
	if data != nil {
		if str, ok := data.(string); ok {
			return c.String(code, str)
		}
	}
	return c.JSON(code, data)
}
func Stm(c echo.Context, buf []byte) error {
	return c.Stream(200, "application/octet-stream", bytes.NewReader(buf))
}

func Uri(c echo.Context) string {
	return c.Scheme() + "://" + c.Request().Host + "/"
}

func QueryParamInt(c echo.Context, key string) int {
	tmp := c.QueryParam(key)
	if n, err := strconv.Atoi(tmp); err == nil {
		return n
	}
	return 0
}

func BindAndValidate(cv echo.Context, i interface{}) error {
	if err := cv.Bind(i); err != nil {
		return err
	}
	if err := Validator.Struct(i); err != nil {
		errMsg := []string{}
		for _, e := range err.(validator.ValidationErrors) {
			errMsg = append(errMsg, e.Translate(trans))
		}
		return errors.New(strings.Join(errMsg, "\n"))
	}
	return nil
}
