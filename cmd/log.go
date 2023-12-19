package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zhiyin2021/zycli/tools"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "log cat, log vi, log ls, log [cmd] yyyyMMdd",
	Long:  `log3 service`,
	Run: func(cmd *cobra.Command, args []string) {
		if logPath := getLogPath(args); logPath != "" {
			cc := exec.Command("tail", "-f", logPath)
			cc.Stdout = os.Stdout
			//异步启动子进程
			cc.Run()
		}
	},
}

var catLogCmd = &cobra.Command{
	Use:   "cat",
	Short: "cat",
	Long:  `cat log `,
	Run: func(cmd *cobra.Command, args []string) {
		if logPath := getLogPath(args); logPath != "" {
			cc := exec.Command("cat", logPath)
			cc.Stdout = os.Stdout
			//异步启动子进程
			cc.Run()
		}
	},
}

var vimLogCmd = &cobra.Command{
	Use:   "vi",
	Short: "vi",
	Long:  `vim log `,
	Run: func(cmd *cobra.Command, args []string) {
		if logPath := getLogPath(args); logPath != "" {
			cc := exec.Command("vi", logPath)
			cc.Stdout = os.Stdout
			cc.Stdin = os.Stdin
			//异步启动子进程
			cc.Run()
		}
	},
}

var lsLogCmd = &cobra.Command{
	Use:   "ls",
	Short: "ls",
	Long:  `ls log `,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("log path:", defOpt.logPath)
		cc := exec.Command("ls", "-lah", defOpt.logPath)
		cc.Stdout = os.Stdout
		cc.Stdin = os.Stdin
		//异步启动子进程
		cc.Run()
	},
}

func getLogPath(args []string) string {
	logName := defOpt.logPath + tools.CurrentName() + ".log"
	if len(args) > 0 {
		if len(args[0]) == 8 {
			logName += "." + args[0]
		} else {
			fmt.Println("View Historical Log Format yyyyMMdd")
			return ""
		}
	}
	if !tools.FileExist(logName) {
		fmt.Println("log file not exist", logName)
		return ""
	}
	return logName
}

type rotateLogsEvent struct{}

func (e *rotateLogsEvent) Handle(ev rotatelogs.Event) {
	if fre, ok := ev.(*rotatelogs.FileRotatedEvent); ok {
		if fre.PreviousFile() != "" {
			logrus.Infof("switch logfile %s => %s", fre.PreviousFile(), fre.CurrentFile())
			go tools.RunCmd("xz", fre.PreviousFile())
		}
	}
}

var logsEv = &rotateLogsEvent{}

func SetLogPath(path string) {
	logPath := path + tools.CurrentName()
	writer3, _ := rotatelogs.New(
		logPath+".dbg.%Y%m%d",                       //每天
		rotatelogs.WithLinkName(logPath+".dbg"),     //生成软链，指向最新日志文件
		rotatelogs.WithRotationTime(10*time.Minute), //最小为5分钟轮询。默认60s  低于1分钟就按1分钟来
		rotatelogs.WithRotationCount(100),           //设置10份 大于10份 或到了清理时间 开始清理
		rotatelogs.WithHandler(logsEv),
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

	writeMap := tools.WriterMap{
		logrus.DebugLevel: writer3,
		logrus.InfoLevel:  writer1,
		logrus.WarnLevel:  writer1,
		logrus.ErrorLevel: writer1,
		logrus.FatalLevel: writer2,
		logrus.PanicLevel: writer2,
	}

	lfHook := tools.NewHook(writeMap, &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "150405.0000",
		// CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
		// 	return f.Function, ""
		// },
	})
	// logrus.SetReportCaller(true)
	logrus.AddHook(lfHook)
}
func init() {
	logCmd.AddCommand(catLogCmd)
	logCmd.AddCommand(vimLogCmd)
	logCmd.AddCommand(lsLogCmd)
	RootCmd.AddCommand(logCmd)
}
