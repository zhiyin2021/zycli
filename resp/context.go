package resp

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/zhiyin2021/zycli/cache"

	"github.com/labstack/echo/v4"
)

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type Context interface {
	echo.Context
	Auth() any
	PageOK(data interface{}, total int64) error
	Ok(data interface{}) error
	ParamErr(msg string, a ...any) error
	BadRequest(msg string, a ...any) error
	NotFound(msg string, a ...any) error
	NoPermission() error
	NoLogin() error
	ServerErr(msg string, a ...any) error
	Json(code int, data interface{}, msg string, a ...any) error
	Resp(code int, data interface{}) error
	Stm(buf []byte) error
	Uri() string
	QueryParamInt(key string) int
	BindAndValidate(i interface{}) error
	Token() string
}

type context struct {
	echo.Context
	auth any
}

var (
	Session  = cache.NewMemory(time.Minute * 30)
	TokenKey = "Authorization"
)

func Ctx(c echo.Context) Context {
	return c.(*context)
}
func (c *context) Auth() any {
	if c.auth == nil {
		token := c.Token()
		if token != "" {
			if item := Session.Get(token); item != nil {
				c.auth = item
			}
		}
	}
	return c.auth
}
func (c *context) Token() string {
	return c.Request().Header.Get(TokenKey)
}
func (c *context) PageOK(data interface{}, total int64) error {
	return c.JSON(http.StatusOK, &Result{Code: 200, Data: H{"total": total, "list": data}})
}

func (c *context) Ok(data interface{}) error {
	return c.JSON(http.StatusOK, &Result{Code: 200, Data: data})
}

func (c *context) ParamErr(msg string, a ...any) error {
	return c.JSON(http.StatusOK, &Result{Code: 400, Msg: "参数解析错误:" + fmt.Sprintf(msg, a...)})
}
func (c *context) BadRequest(msg string, a ...any) error {
	return c.JSON(http.StatusOK, &Result{Code: 400, Msg: fmt.Sprintf(msg, a...)})
}
func (c *context) NotFound(msg string, a ...any) error {
	return c.JSON(http.StatusOK, &Result{Code: 404, Msg: fmt.Sprintf(msg, a...)})
}

func (c *context) NoPermission() error {
	return c.JSON(http.StatusOK, &Result{Code: 403})
}

func (c *context) NoLogin() error {
	return c.JSON(http.StatusUnauthorized, &Result{Code: 401})
}

func (c *context) ServerErr(msg string, a ...any) error {
	return c.JSON(http.StatusOK, &Result{Code: 500, Msg: fmt.Sprintf(msg, a...)})
}
func (c *context) Json(code int, data interface{}, msg string, a ...any) error {
	return c.JSON(http.StatusOK, &Result{Code: code, Data: data, Msg: fmt.Sprintf(msg, a...)})
}
func (c *context) Resp(code int, data interface{}) error {
	if data != nil {
		if str, ok := data.(string); ok {
			return c.String(code, str)
		}
	}
	return c.JSON(code, data)
}
func (c *context) Stm(buf []byte) error {
	return c.Stream(200, "application/octet-stream", bytes.NewReader(buf))
}

func (c *context) Uri() string {
	return c.Context.Scheme() + "://" + c.Request().Host + "/"
}

func (c *context) QueryParamInt(key string) int {
	tmp := c.QueryParam(key)
	if n, err := strconv.Atoi(tmp); err == nil {
		return n
	}
	return 0
}

func (cv *context) BindAndValidate(i interface{}) error {
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
