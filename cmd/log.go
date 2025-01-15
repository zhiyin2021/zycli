package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/zhiyin2021/zycli/tools"
	"github.com/zhiyin2021/zycli/tools/logger"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "log cat,log ls, log [cmd] yyyyMMdd",
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
			if !tools.FileExists(logPath) {
				cc := exec.Command("zcat", logPath+".gz")
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
	if !tools.FileExists(logName) {
		fmt.Println("log file not exist", logName)
		return logName, errors.New("log file not exists " + logName)
	}
	return logName, nil
}

func (opt *cmdOpt) initLog() {
	logPath := opt.logPath + tools.CurrentName()

	logWrite := NewSplit(logPath+".log", func(l *logWriter) {
		l.maxAge = opt.maxAge
		l.maxCount = opt.maxCount
		l.compressType = opt.compressType
	}, OptMaxSize(opt.maxSize))

	logger.SetLogger(logWrite)
}
func init() {
	logCmd.AddCommand(catLogCmd)
	logCmd.AddCommand(lsLogCmd)
	RootCmd.AddCommand(logCmd)
}
