package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zhiyin2021/zycli/tools"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "log cat, log vi, log ls, log [cmd] yyyyMMdd",
	Long:  `log3 service`,
	Run: func(cmd *cobra.Command, args []string) {
		if logPath, err := getLogPath(args); err == nil {
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
		if logPath, _ := getLogPath(args); logPath != "" {
			if !tools.FileExist(logPath) {
				cc := exec.Command("xzcat", logPath+".xz")
				cc.Stdout = os.Stdout
				//异步启动子进程
				cc.Run()
			} else {
				cc := exec.Command("cat", logPath)
				cc.Stdout = os.Stdout
				//异步启动子进程
				cc.Run()
			}
		}
	},
}

var vimLogCmd = &cobra.Command{
	Use:   "vi",
	Short: "vi",
	Long:  `vim log `,
	Run: func(cmd *cobra.Command, args []string) {
		if logPath, err := getLogPath(args); err == nil {
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

func getLogPath(args []string) (string, error) {
	logName := defOpt.logPath + tools.CurrentName() + ".log"
	if len(args) > 0 {
		if len(args[0]) == 8 {
			logName += "." + args[0]
		} else {
			fmt.Println("View Historical Log Format yyyyMMdd")
			return logName, errors.New("view Historical Log Format yyyyMMdd")
		}
	}
	if !tools.FileExist(logName) {
		fmt.Println("log file not exist", logName)
		return logName, errors.New("log file not exists " + logName)
	}
	return logName, nil
}

func (opt *cmdOpt) initLog() {
	logPath := opt.logPath + tools.CurrentName()
	dbgWrite := NewSplit(logPath+".dbg", func(l *logWriter) {
		l.maxAge = opt.maxAge
		l.maxCount = opt.maxCount
		l.compressType = opt.compressType
	}, OptMaxSize(opt.maxSize))
	logWrite := NewSplit(logPath+".log", func(l *logWriter) {
		l.maxAge = opt.maxAge
		l.maxCount = opt.maxCount
		l.compressType = opt.compressType
	}, OptMaxSize(opt.maxSize))

	// logrus.SetOutput(logWrite)
	writeMap := tools.WriterMap{
		logrus.DebugLevel: dbgWrite,
		logrus.InfoLevel:  logWrite,
		logrus.WarnLevel:  logWrite,
		logrus.ErrorLevel: logWrite,
		logrus.FatalLevel: logWrite,
		logrus.PanicLevel: logWrite,
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
