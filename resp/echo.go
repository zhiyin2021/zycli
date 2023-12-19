package resp

import (
	cc "context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"regexp"
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

type (
	H              map[string]any
	HandlerFunc    func(c Context) error
	MiddlewareFunc func(next HandlerFunc) HandlerFunc
	echoext        struct {
		echo *echo.Echo
	}
)

func init() {

	// jsoniter.RegisterTypeDecoderFunc("time.Time", func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	// 	t, err := time.ParseInLocation("2006-01-02 15:04:05", iter.ReadString(), time.UTC)
	// 	if err != nil {
	// 		iter.Error = err
	// 		return
	// 	}
	// 	*((*time.Time)(ptr)) = t
	// })

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
		_echo.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			Skipper:      middleware.DefaultSkipper,
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		}))

		_echo.echo.JSONSerializer = &JSONSerializer{}
		_echo.echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				cc := &context{c, nil, validator.New()}

				if err := next(cc); err != nil {
					logrus.Errorf("error=>%s=>%s=>%s", err, c.Request().Method, c.Request().URL.Path)
					if he, ok := err.(*echo.HTTPError); ok {
						message := fmt.Sprintf("%v", he.Message)
						return c.JSON(200, H{"code": he.Code, "msg": message})
					}
					code := 500
					if err.Error() == "record not found" {
						log.Println("c.Response().Status", c.Response().Status)
						code = 404
					}
					return c.JSON(200, H{"code": code, "msg": err.Error()})
				}
				if c.Response().Status == 404 {
					logrus.Errorf("succ1=>%d=>%s=>%s", c.Response().Status, c.Request().Method, c.Request().URL.Path)
					c.Response().Status = 200
					return c.HTMLBlob(200, []byte{})
				}
				return nil
			}
		})
		// _echo.Use(auth)
	})
	return _echo
}

func (e *echoext) StaticFS(wwwFS embed.FS, indexPath string) {
	if indexPath != "" {
		buf, _ := wwwFS.ReadFile(indexPath)
		_echo.routeNotFound("/*", func(c echo.Context) error {
			if c.Request().Header.Get("content-type") == "application/json" {
				return c.JSON(200, H{"code": 404, "msg": "not found"})
			} else if ok, _ := regexp.MatchString(`/.*/`, c.Request().URL.Path); ok {
				return nil
			} else {
				return c.HTMLBlob(200, buf)
			}
		})
	}
	assetHandler := http.FileServer(getFS(wwwFS))
	_echo.GET("/", wrapHandler(assetHandler))
	_echo.GET("/assets/*", wrapHandler(http.StripPrefix("/", assetHandler)))
}
func (e *echoext) Static(path string, root string) *echo.Route {
	return _echo.echo.Static(path, root)
}
func getFS(embedFS embed.FS) http.FileSystem {
	useOS := len(os.Args) > 1 && os.Args[1] == "live"
	if useOS {
		log.Print("using live mode")
		return http.FS(os.DirFS("dist"))
	}
	fsys, err := fs.Sub(embedFS, "dist")
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}
func (e *echoext) Group(path string, m ...MiddlewareFunc) *Group {
	return &Group{e.echo.Group(path, toMiddle(m...)...)}
}
func toMiddle(mws ...MiddlewareFunc) []echo.MiddlewareFunc {
	var middleware []echo.MiddlewareFunc
	for _, mw := range mws {
		middleware = append(middleware, func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				next1 := func(c Context) error {
					return next(c.(*context))
				}
				return mw(next1)(c.(*context))
			}
		})
	}
	return middleware
}
func (e *echoext) Use(m ...MiddlewareFunc) {
	e.echo.Use(toMiddle(m...)...)
}
func (e *echoext) routeNotFound(path string, h echo.HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.echo.RouteNotFound(path, h, toMiddle(m...)...)
}
func (e *echoext) GET(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.echo.GET(path, func(c echo.Context) error {
		return h(c.(*context))
	}, toMiddle(m...)...)
}
func (e *echoext) POST(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.echo.POST(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
}

func (e *echoext) PUT(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.echo.PUT(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
}

func (e *echoext) PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.echo.PATCH(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
}
func (e *echoext) OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.echo.OPTIONS(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
}
func (e *echoext) DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.echo.DELETE(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
}
func (e *echoext) HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.echo.HEAD(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
}
func (e *echoext) TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.echo.TRACE(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
}
func (e *echoext) CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.echo.CONNECT(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
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
func (e *echoext) StartTLS(addr string, crtPEM, keyPEM any) error {
	return e.echo.Start(addr)
}
func (e *echoext) Shutdown(ctx cc.Context) error {
	return e.echo.Shutdown(ctx)
}

type Group struct {
	group *echo.Group
}

func (e *Group) Use(m ...MiddlewareFunc) {
	e.group.Use(toMiddle(m...)...)
}

func (e *Group) GET(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.group.GET(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
}
func (e *Group) POST(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.group.POST(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
}

func (e *Group) PUT(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.group.PUT(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
}

func (e *Group) PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.group.PATCH(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
}
func (e *Group) OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.group.OPTIONS(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
}
func (e *Group) DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.group.DELETE(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
}
func (e *Group) HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.group.HEAD(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
}
func (e *Group) TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.group.TRACE(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
}
func (e *Group) CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) *echo.Route {
	return e.group.CONNECT(path, func(c echo.Context) error { return h(c.(*context)) }, toMiddle(m...)...)
}
