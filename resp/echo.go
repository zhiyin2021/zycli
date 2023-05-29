package resp

import (
	cc "context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/go-playground/validator"
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

var (
	runOnce sync.Once
	_echo   *echoext
)

type HandlerFunc func(c Context) error
type echoext struct {
	echo *echo.Echo
}

func init() {
	jsoniter.RegisterTypeEncoderFunc("time.Time", func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
		t := *((*time.Time)(ptr))
		stream.WriteString(t.Format("2006-01-02 15:04:05"))
	}, func(ptr unsafe.Pointer) bool {
		return false
	})
}
func GetEcho() *echoext {
	runOnce.Do(func() {
		_echo = &echoext{echo: echo.New()}
		_echo.echo.HideBanner = true
		// e.Use(middleware.Recover())
		_echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			Skipper:      middleware.DefaultSkipper,
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		}))
		_echo.echo.JSONSerializer = &JSONSerializer{}
		_echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				if err := next(c); err != nil {
					logrus.Errorf("error=>%s=>%s=>%s", err, c.Request().Method, c.Request().URL.Path)
					if he, ok := err.(*echo.HTTPError); ok {
						message := fmt.Sprintf("%v", he.Message)
						return c.JSON(200, map[string]interface{}{"code": he.Code, "msg": message})
					}
					code := 500
					if err.Error() == "record not found" {
						log.Println("c.Response().Status", c.Response().Status)
						code = 404
					}
					return c.JSON(200, map[string]interface{}{"code": code, "msg": err.Error()})
				}
				if c.Response().Status == 404 {
					logrus.Errorf("succ1=>%d=>%s=>%s", c.Response().Status, c.Request().Method, c.Request().URL.Path)
					c.Response().Status = 200
					return c.HTMLBlob(200, []byte{})
				}
				return nil
			}
		})
		_echo.Use(auth)
	})
	return _echo
}

func (e *echoext) Static(www http.FileSystem, indexFile []byte) {
	if indexFile != nil {
		_echo.routeNotFound("/*", func(c echo.Context) error {
			if strings.HasPrefix(c.Request().URL.Path, "/api") {
				return c.JSON(200, map[string]interface{}{"code": 404, "msg": "not found"})
			} else {
				return c.HTMLBlob(200, indexFile)
			}
		})
	}
	assetHandler := http.FileServer(www)
	_echo.GET("/", wrapHandler(assetHandler))
	_echo.GET("/assets/*", wrapHandler(http.StripPrefix("/", assetHandler)))
}

func (e *echoext) Group(path string) *Group {
	return &Group{e.echo.Group(path)}
}
func (e *echoext) Use(middleware ...echo.MiddlewareFunc) {
	e.echo.Use(middleware...)
}
func (e *echoext) routeNotFound(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return e.echo.RouteNotFound(path, h, m...)
}
func (e *echoext) GET(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return e.echo.GET(path, func(c echo.Context) error { return h(c.(*context)) }, m...)
}
func (e *echoext) POST(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return e.echo.POST(path, func(c echo.Context) error { return h(c.(*context)) }, m...)
}
func wrapHandler(h http.Handler) HandlerFunc {
	return func(c Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
func (e *echoext) Start(addr string) error {
	return e.echo.Start(addr)
}
func (e *echoext) Shutdown(ctx cc.Context) error {
	return e.echo.Shutdown(ctx)
}

type Group struct {
	group *echo.Group
}

func (e *Group) GET(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return e.group.GET(path, func(c echo.Context) error { return h(c.(*context)) }, m...)
}
func (e *Group) POST(path string, h HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route {
	return e.group.POST(path, func(c echo.Context) error { return h(c.(*context)) }, m...)
}

var AnonymousUrls = []string{"/api/user.login", "/api/login"}

func auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc := &context{c, nil, validator.New()}
		uri := c.Request().RequestURI
		if strings.HasPrefix(uri, "/api") {
			// 路由拦截 - 登录身份、资源权限判断等
			for i := range AnonymousUrls {
				if strings.HasPrefix(uri, AnonymousUrls[i]) {
					return next(cc)
				}
			}
			token := cc.Request().Header.Get("Authorization")
			if token != "" {
				if item := goCahce.Get(token); item != nil {
					cc.auth = item.(*AuthInfo)
				}
			}
			if cc.auth == nil {
				logrus.Warnf("401 [%s] %s", uri, token)
				return cc.NoLogin()
			}
			// authorization := v.(dto.Authorization)
			// if strings.EqualFold(constant.LoginToken, authorization.Type) {
			// 	if authorization.Remember {
			// 		// 记住登录有效期两周
			// 		cache.TokenManager.Set(token, authorization, cache.RememberMeExpiration)
			// 	} else {
			// 		cache.TokenManager.Set(token, authorization, cache.NotRememberExpiration)
			// 	}
			// }
			return next(cc)
		}
		return next(cc)
	}
}
