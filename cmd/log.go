package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"

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
		if logPath := getLogPath(args); logPath != "" {
			logPath = path.Dir(logPath)
			cc := exec.Command("ls", "-lah", logPath)
			cc.Stdout = os.Stdout
			cc.Stdin = os.Stdin
			//异步启动子进程
			cc.Run()
		}
	},
}

func getLogPath(args []string) string {
	logName := os.Getenv("ZYCLI_" + tools.CurrentName() + "_LOG")
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

func init() {
	logCmd.AddCommand(catLogCmd)
	logCmd.AddCommand(vimLogCmd)
	logCmd.AddCommand(lsLogCmd)
	RootCmd.AddCommand(logCmd)
}
