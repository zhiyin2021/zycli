## gocli

> cmd => cobra 二次封装
> cache => 缓存类扩展
> resp => ECHO 自定义扩展
> tool => 工具类

### step 1

```shell
go get -u github.com/zhiyin2021/zycli
```

### main.go

```golang
package main

import (
	"context"
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
)

type Config struct {
	ConnStr string `json:"connStr"`
	Port    int    `json:"port"`
}

var config Config

func main() {
	cmd.Execute(func(cmd *cobra.Command, args []string) {
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
	}, true)
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


```

```shell
#启动程序
app
#安装服务
app install
#卸载服务
app uninstall
#启动服务
app start
#停止服务
app stop
#tail -f app.log 方式查看最近日志
app log
#cat app.log 方式查看日志
app log cat
```
