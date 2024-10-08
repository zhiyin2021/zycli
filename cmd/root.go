package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zhiyin2021/zycli/tools"
)

type cmdOpt struct {
	regSvc  bool
	logPath string
	ipcPath string
	// 日志相关
	maxSize      int64
	maxAge       int
	maxCount     int
	compressType CompressType
	layout       string
}

type Option func(*cmdOpt)

var (
	Version   = "0.0.1"
	DEBUG     = false
	svcFunc   func([]string)
	quit, sig = make(chan os.Signal), make(chan os.Signal, 1)
	defOpt    = &cmdOpt{
		regSvc:       false,
		logPath:      tools.CurrentDir() + "/log/",
		ipcPath:      tools.FixPath(tools.CurrentName() + ".ipc"),
		maxSize:      1000,
		maxAge:       90,
		maxCount:     0,
		compressType: CT_XZ,
		layout:       "060102_150405_000",
	}
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
				logrus.Println("wait quit")
				wg.Wait()
			case <-time.After(10 * time.Millisecond):
				logrus.Println("system quit")
			}
		}
	},
}

func WithLogPath(path string) Option {
	return func(opt *cmdOpt) {
		opt.logPath = path
	}
}

func WithRegSvc() Option {
	return func(opt *cmdOpt) {
		opt.regSvc = true
	}
}

func WithIpcPath(ipcPath string) Option {
	return func(opt *cmdOpt) {
		opt.ipcPath = ipcPath
	}
}
func WithLogMaxSize(maxSize int64) Option {
	return func(opt *cmdOpt) {
		opt.maxSize = maxSize
	}
}
func WithLogMaxAge(maxAge int) Option {
	return func(opt *cmdOpt) {
		opt.maxAge = maxAge
	}
}
func WithLogMaxCount(maxCount int) Option {
	return func(opt *cmdOpt) {
		opt.maxCount = maxCount
	}
}
func WithLogCompressType(compressType CompressType) Option {
	return func(opt *cmdOpt) {
		opt.compressType = compressType
	}
}
func WithLogLayout(layout string) Option {
	return func(opt *cmdOpt) {
		opt.layout = layout
	}
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
	logrus.SetOutput(os.Stdout)
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
	// SetLogPath(defOpt.logPath)
	defOpt.initLog()

	file, err := os.OpenFile("panic.err", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法打开文件: %v\n", err)
		return
	}
	defer file.Close()

	// 使用 syscall 重定向标准错误输出
	if err := syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd())); err != nil {
		fmt.Fprintf(os.Stderr, "重定向错误输出失败: %v\n", err)
		return
	}

	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "root.execute:"+err.Error())
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
