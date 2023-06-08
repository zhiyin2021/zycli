package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
func run(ccmd *cobra.Command, args []string) {
	initConfig()
	e := resp.GetEcho()
	addr := fmt.Sprintf("0.0.0.0:%d", config.Port)
	logrus.Println("server start at ", addr)
	go e.Start(addr)
	e.GET("/", helloworld)
	code := <-cmd.WaitQuit()
	logrus.Println("server stop", code)
	ctx, cancel := context.WithCancel(context.Background())
	e.Shutdown(ctx)
	cancel()
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
