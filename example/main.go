package main

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zhiyin2021/zycli/cmd"
	"github.com/zhiyin2021/zycli/resp"
	"github.com/zhiyin2021/zycli/tools"
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
	e := resp.GetEcho()
	addr := fmt.Sprintf("0.0.0.0:%d", config.Port)
	logrus.Println("server start at ", addr)

	e.GET("/", helloworld)
	e.GET("/test", testPanic)
	go e.Start(addr)
}
func helloworld(ctx resp.Context) error {
	if cmd.DEBUG {
		logrus.Debugln("hello world, debug mode")
		return ctx.String(200, "hello world, debug mode")
	}
	logrus.Infoln("hello world")
	return ctx.String(200, "hello world")
}
func testPanic(ctx resp.Context) error {
	go func() {
		time.Sleep(500 * time.Millisecond)
		panic("test panic")
	}()
	return ctx.String(200, "hello world")
}
func initConfig() {
	var err error
	config, err = tools.LoadConfig[Config]("config.json", json.Unmarshal)
	if err != nil {
		logrus.Warnln("load config", err)
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
				logrus.Errorf("%v\n%s", err, debug.Stack())
			}
		}()
		f()
	}()
}
func Try(f func()) {
	defer func() {
		err := recover()
		if err != nil {
			logrus.Errorf("%v\n%s", err, debug.Stack())
		}
	}()
	f()
}
