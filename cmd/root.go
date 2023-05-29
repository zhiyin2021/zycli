package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zhiyin2021/zycli/tools"
)

var (
	Version = "0.0.1"
	DEBUG   = false
)
var RootCmd = &cobra.Command{
	Use:     tools.CurrentName(),
	Short:   tools.CurrentName() + " server.",
	Long:    tools.CurrentName() + ` server.`,
	Version: Version,
}

func Execute() {
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
	if DEBUG {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&DEBUG, "debug", false, "start with debug mode")
}
