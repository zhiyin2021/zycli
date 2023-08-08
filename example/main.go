package main

import (
	"context"
	"encoding/json"
	"fmt"

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
	cmd.Execute(run, false)
}
func run(args []string) {

	initConfig()
	e := resp.GetEcho()
	addr := fmt.Sprintf("0.0.0.0:%d", config.Port)
	logrus.Println("server start at ", addr)

	e.GET("/", helloworld)
	go e.Start(addr)
	<-cmd.WaitQuit()
	logrus.Println("server stop1")
	ctx, cancel := context.WithCancel(context.Background())
	logrus.Println("server stop2")
	e.Shutdown(ctx)
	logrus.Println("server stop3")
	cancel()
	logrus.Println("server stop4")
}
func helloworld(ctx resp.Context) error {
	if cmd.DEBUG {
		logrus.Debugln("hello world, debug mode")
		return ctx.String(200, "hello world, debug mode")
	}
	logrus.Infoln("hello world")
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
