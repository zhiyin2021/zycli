package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zhiyin2021/zycli/tools"
)

var (
	Version = "0.0.1"
	DEBUG   = false
)
var RootCmd = &cobra.Command{
	Use:     tools.CurrentName(),
	Short:   "freeswitch asr server.",
	Long:    `freeswitch asr server.`,
	Version: Version,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&DEBUG, "debug", false, "start with debug mode")
}
