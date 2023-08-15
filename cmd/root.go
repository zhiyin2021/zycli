package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zhiyin2021/zycli/tools"
)

var (
	Version   = "0.0.1"
	DEBUG     = false
	svcFunc   func([]string)
	quit, sig = make(chan os.Signal), make(chan os.Signal)
)
var RootCmd = &cobra.Command{
	Use:     tools.CurrentName(),
	Short:   tools.CurrentName() + " server.",
	Long:    tools.CurrentName() + ` server.`,
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		if svcFunc != nil {
			logName := tools.LogPath() + tools.CurrentName() + ".log"
			writer, _ := rotatelogs.New(
				logName+".%Y%m%d",                           //每天
				rotatelogs.WithLinkName(logName),            //生成软链，指向最新日志文件
				rotatelogs.WithRotationTime(10*time.Minute), //最小为5分钟轮询。默认60s  低于1分钟就按1分钟来
				rotatelogs.WithRotationCount(10),            //设置10份 大于10份 或到了清理时间 开始清理
				rotatelogs.WithRotationSize(256*1024*1024),  //设置100MB大小,当大于这个容量时，创建新的日志文件
			)
			mw := io.MultiWriter(os.Stdout, writer)
			logrus.SetOutput(mw)
			if !DEBUG {
				DEBUG = tools.FileExist(tools.CurrentName() + ".dbg")
			}
			err := startUnixSock()
			if err != nil {
				return
			}
			defer stopUnixSock()

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer func() {
					if err := recover(); err != nil {
						fmt.Println("panic recoverd ==> ", err)
						fmt.Println("stack ==> ", string(debug.Stack()))
						sig <- syscall.SIGTERM
					}
					wg.Done()
				}()
				svcFunc(args)
			}()

			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
			s := <-sig
			select {
			case quit <- s:
				log.Println("wait quit")
				wg.Wait()
			case <-time.After(10 * time.Millisecond):
				log.Println("system quit")
			}
		}
	},
}

func WaitQuit() <-chan os.Signal {
	return quit
}
func Quit() {
	sig <- syscall.SIGQUIT
}

// mainFunc 主函数
// regSvc 是否注册服务
func Execute(mainFunc func([]string), regSvc bool) {
	if DEBUG {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if mainFunc == nil {
		panic("MainFunc is nil")
	}
	svcFunc = mainFunc
	if regSvc {
		addSvc()
	}

	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "log",
	Long:  `log service`,
	Run: func(cmd *cobra.Command, args []string) {
		fname := tools.CurrentName()
		var cc *exec.Cmd
		if len(args) > 0 && args[0] == "cat" {
			cc = exec.Command("cat", "status", tools.LogPath()+fname+".log")
		} else {
			cc = exec.Command("tail", "-f", tools.LogPath()+fname+".log")
		}
		cc.Stdout = os.Stdout
		//异步启动子进程
		cc.Run()
	},
}

var dbgCmd = &cobra.Command{
	Use:   "dbg",
	Short: "dbg",
	Long:  `enabled debug`,
	Run: func(cmd *cobra.Command, args []string) {
		msg, err := SendMsgToIPC("dbg")
		if err != nil {
			logrus.Errorln("please check application not running.")
		} else {
			logrus.Infoln(msg)
		}
	},
}

// var serverCmd = &cobra.Command{
// 	Use:   "server",
// 	Short: "Start the server at the specified address",
// 	Long: `Start the server at the specified address
// the address is defined in config file`,
// 	Run: svcFunc,
// }

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "150405.0000", //时间格式化
	})
	RootCmd.PersistentFlags().BoolVar(&DEBUG, "debug", false, "start with debug mode")
	RootCmd.AddCommand(logCmd)
	RootCmd.AddCommand(dbgCmd)
}
