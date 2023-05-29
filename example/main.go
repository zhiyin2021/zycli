package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zhiyin2021/zycli/cmd"
	"github.com/zhiyin2021/zycli/resp"
	"github.com/zhiyin2021/zycli/tools"
	"golang.org/x/net/context"
)

type Config struct {
	ConnStr string `json:"connStr"`
	Port    int    `json:"port"`
}

var config Config

func main() {
	cmd.Execute()
}

// ServerCmd represents the server command
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server at the specified address",
	Long: `Start the server at the specified address
the address is defined in config file`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()
		e := resp.GetEcho()
		addr := fmt.Sprintf("0.0.0.0:%d", config.Port)
		logrus.Println("server start at ", addr)
		go e.Start(addr)
		e.GET("/", helloworld)
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		ctx, cancel := context.WithCancel(context.Background())
		e.Shutdown(ctx)
		cancel()
	},
}

func init() {
	cmd.RootCmd.AddCommand(ServerCmd)
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
