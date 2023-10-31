package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
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

type cmdOpt struct {
	regSvc  bool
	logPath string
	ipcPath string
}

type Option func(*cmdOpt)

var (
	Version   = "0.0.1"
	DEBUG     = false
	svcFunc   func([]string)
	quit, sig = make(chan os.Signal), make(chan os.Signal)
	defOpt    = &cmdOpt{regSvc: false, logPath: tools.CurrentDir() + "/log/", ipcPath: tools.FixPath(tools.CurrentName() + ".ipc")}
)
var RootCmd = &cobra.Command{
	Use:     tools.CurrentName(),
	Short:   tools.CurrentName() + " server.",
	Long:    tools.CurrentName() + ` server.`,
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("--------------------\n  app: %s \n  ver: %s \n--------------------\n", tools.CurrentName(), Version)
		if svcFunc != nil {
			if !DEBUG {
				DEBUG = tools.FileExist(tools.CurrentName() + ".dbg")
			}
			if IsRuning() {
				fmt.Println("already running")
				return
			}
			err := startUnixSock()
			if err != nil {
				return
			}
			defer stopUnixSock()

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer OnPanic(func(a any, s string) {
					sig <- syscall.SIGTERM
				})
				defer wg.Done()
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

func WithLogPath(path string) Option {
	return func(opt *cmdOpt) {
		opt.logPath = path
	}
}

func WithRegSvc(reged bool) Option {
	return func(opt *cmdOpt) {
		opt.regSvc = reged
	}
}

func WithIpcPath(ipcPath string) Option {
	return func(opt *cmdOpt) {
		opt.ipcPath = ipcPath
	}
}
func SetLogPath(path string) {
	logPath := path + tools.CurrentName()
	writer3, _ := rotatelogs.New(
		logPath+".dbg.%Y%m%d",                       //每天
		rotatelogs.WithLinkName(logPath+".dbg"),     //生成软链，指向最新日志文件
		rotatelogs.WithRotationTime(10*time.Minute), //最小为5分钟轮询。默认60s  低于1分钟就按1分钟来
		rotatelogs.WithRotationCount(100),           //设置10份 大于10份 或到了清理时间 开始清理
	)

	writer1, _ := rotatelogs.New(
		logPath+".log.%Y%m%d",                       //每天
		rotatelogs.WithLinkName(logPath+".log"),     //生成软链，指向最新日志文件
		rotatelogs.WithRotationTime(10*time.Minute), //最小为5分钟轮询。默认60s  低于1分钟就按1分钟来
		rotatelogs.WithRotationCount(100),           //设置10份 大于10份 或到了清理时间 开始清理
	)

	writer2, _ := rotatelogs.New(
		logPath+".err.%Y%m%d",                       //每天
		rotatelogs.WithLinkName(logPath+".err"),     //生成软链，指向最新日志文件
		rotatelogs.WithRotationTime(10*time.Minute), //最小为5分钟轮询。默认60s  低于1分钟就按1分钟来
		rotatelogs.WithRotationCount(100),           //设置10份 大于10份 或到了清理时间 开始清理
	)

	mw1 := io.MultiWriter(os.Stdout, writer1)
	mw2 := io.MultiWriter(os.Stdout, writer2)
	mw3 := io.MultiWriter(os.Stdout, writer3)
	writeMap := tools.WriterMap{
		logrus.DebugLevel: mw3,
		logrus.InfoLevel:  mw1,
		logrus.WarnLevel:  mw1,
		logrus.ErrorLevel: mw2,
	}
	lfHook := tools.NewHook(writeMap, &logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "150405.0000",
	})
	logrus.AddHook(lfHook)
}
func WaitQuit() <-chan os.Signal {
	return quit
}
func Quit() {
	sig <- syscall.SIGQUIT
}

// mainFunc 主函数
// regSvc 是否注册服务
func Execute(mainFunc func([]string), opts ...Option) {
	if DEBUG {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if mainFunc == nil {
		panic("MainFunc is nil")
	}
	for _, opt := range opts {
		opt(defOpt)
	}
	if defOpt.logPath == "" {
		defOpt.logPath = tools.CurrentDir() + "/log/"
	}
	svcFunc = mainFunc
	if defOpt.regSvc {
		addSvc()
	}
	SetLogPath(defOpt.logPath)

	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var dbgCmd = &cobra.Command{
	Use:   "dbg",
	Short: "dbg",
	Long:  `enabled debug`,
	Run: func(cmd *cobra.Command, args []string) {
		msg, err := SendMsgToIPC("dbg")
		if err != nil {
			if err.Error() != "EOF" {
				logrus.Errorln("please check application not running:", err)
			}
		} else {
			logrus.Infoln(msg)
		}
	},
}

func init() {
	logFmt := &logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "150405.0000", //时间格式化
	}
	logrus.SetFormatter(logFmt)
	RootCmd.PersistentFlags().BoolVar(&DEBUG, "debug", false, "start with debug mode")
	RootCmd.AddCommand(dbgCmd)
}

func OnPanic(call func(any, string)) {
	if err := recover(); err != nil {
		stack := string(debug.Stack())
		logrus.Panic(err, "\n", stack)
		if call != nil {
			call(err, stack)
		}
	}
}
