package main

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/zhiyin2021/zycli/cmd"
	"github.com/zhiyin2021/zycli/resp"
	"github.com/zhiyin2021/zycli/tools"
	"github.com/zhiyin2021/zycli/tools/logger"
)

type Config struct {
	ConnStr string `json:"connStr"`
	Port    int    `json:"port"`
}

var config Config

func main() {
	cmd.Execute(run, cmd.WithRegSvc())
}

func run(args []string) {
	initConfig()
	e := resp.Server()
	addr := fmt.Sprintf("0.0.0.0:%d", config.Port)
	logger.Println("server start at ", addr)

	e.GET("/", helloworld)
	e.GET("/test", testPanic)
	go e.Start(addr)
}

func helloworld(ctx echo.Context) error {
	if cmd.DEBUG {
		logger.Debugln("hello world, debug mode")
		return ctx.String(200, "hello world, debug mode")
	}
	logger.Infoln("hello world")
	return resp.BadRequest(ctx, "hello world")
}
func testPanic(ctx echo.Context) error {
	go func() {
		time.Sleep(500 * time.Millisecond)
		panic("test panic")
	}()
	return resp.Ok(ctx, "hello world")
}
func initConfig() {
	var err error
	config, err = tools.LoadConfig[Config]("config.json", json.Unmarshal)
	if err != nil {
		logger.Warnln("load config", err)
		config = Config{
			Port: 8080,
		}
	}
}

func TryGO(f func()) {
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				logger.Errorf("%v\n%s", err, debug.Stack())
			}
		}()
		f()
	}()
}
func Try(f func()) {
	defer func() {
		err := recover()
		if err != nil {
			logger.Errorf("%v\n%s", err, debug.Stack())
		}
	}()
	f()
}
