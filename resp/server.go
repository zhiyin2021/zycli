package resp

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"sync"
	"time"
	"unsafe"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/zhiyin2021/zycli/tools/logger"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	zhTrans "github.com/go-playground/validator/v10/translations/zh"
)

var (
	runOnce   sync.Once
	server    *echo.Echo
	Validator *validator.Validate
	trans     ut.Translator
)

type (
	H map[string]any
)

func init() {
	Validator = validator.New()
	uniTrans := ut.New(zh.New())
	trans, _ = uniTrans.GetTranslator("zh")
	// 注册翻译器到验证器
	err := zhTrans.RegisterDefaultTranslations(Validator, trans)
	if err != nil {
		panic(fmt.Sprintf("registerDefaultTranslations fail: %s\n", err.Error()))
	}
	Validator.RegisterTagNameFunc(func(field reflect.StructField) string {
		label := field.Tag.Get("label")
		if label == "" {
			return field.Name
		}
		return label
	})
	jsoniter.RegisterTypeEncoderFunc("time.Time", func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
		t := *((*time.Time)(ptr))
		stream.WriteString(t.Format("2006-01-02 15:04:05"))
	}, func(ptr unsafe.Pointer) bool {
		return false
	})
}
func Server() *echo.Echo {
	runOnce.Do(func() {
		server = echo.New()
		server.HideBanner = true
		server.JSONSerializer = &JSONSerializer{}
		server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			Skipper:      middleware.DefaultSkipper,
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		}))
		server.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			defer func() {
				if err := recover(); err != nil {
					if err == http.ErrAbortHandler {
						panic(err)
					}
					logger.Errorln("http.error:", err)
				}
			}()
			return func(c echo.Context) error {
				return next(c)
			}
		})
	})
	return server
}

func StaticFS(wwwFS embed.FS, indexPath string) {
	if indexPath != "" {
		buf, _ := wwwFS.ReadFile(indexPath)
		server.RouteNotFound("/*", func(c echo.Context) error {
			if c.Request().Header.Get("content-type") == "application/json" {
				return c.JSON(200, Result{Code: 404, Msg: "notfound"})
			} else if ok, _ := regexp.MatchString(`/.*/`, c.Request().URL.Path); ok {
				return c.NoContent(200)
			} else {
				return c.HTMLBlob(200, buf)
			}
		})
	}
	assetHandler := http.FileServer(getFS(wwwFS))
	server.GET("/", wrapHandler(assetHandler))
	server.GET("/assets/*", wrapHandler(http.StripPrefix("/", assetHandler)))
}
func Static(path string, root string) *echo.Route {
	return server.Static(path, root)
}
func wrapHandler(h http.Handler) echo.HandlerFunc {
	return func(c echo.Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
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
