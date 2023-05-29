package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/zhiyin2021/zycli/tools"
)

var (
	daemoCmd *exec.Cmd
)

// ServerCmd represents the server command
var svcCmd = &cobra.Command{
	Use:   "nc",
	Short: "backgroud service",
	Long:  `backgroud Service`,
	Run: func(cmd *cobra.Command, args []string) {
		daemoCmd = exec.Command(os.Args[0], "server", "log")
		//异步启动子进程
		err := daemoCmd.Start()
		if err != nil {
			panic(err)
		}
		log.Println("backgroup started => ", daemoCmd.Process.Pid)
	},
}
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install",
	Long:  `backgroud service`,
	Run: func(cmd *cobra.Command, args []string) {
		fname := tools.CurrentName()
		path := tools.CurrentDir()
		data := fmt.Sprintf(svcText, fname, path, path, fname)

		os.Remove("/usr/sbin/" + fname)
		os.Symlink(path+"/"+fname, "/usr/sbin/"+fname)

		err := os.WriteFile("/etc/systemd/system/"+fname+".service", []byte(data), 0644)
		if err == nil {
			log.Println("generate service success")
			if err = run("systemctl", "daemon-reload"); err == nil {
				log.Println("reload service success")
				ctlSvc("enable")
				if err = ctlSvc("start"); err == nil {
					return
				}
			}
		}
		log.Println("service install error", err)
	},
}
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "uninstall",
	Long:  `backgroud service`,
	Run: func(cmd *cobra.Command, args []string) {
		fname := tools.CurrentName()
		ctlSvc("stop")
		ctlSvc("disable")
		os.Remove("/etc/systemd/system/" + fname + ".service")
		run("systemctl", "daemon-reload")
		os.Remove("/usr/sbin/" + fname)
	},
}
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start",
	Long:  `start service`,
	Run: func(cmd *cobra.Command, args []string) {
		ctlSvc("start")
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop",
	Long:  `stop service`,
	Run: func(cmd *cobra.Command, args []string) {
		ctlSvc("stop")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "status",
	Long:  `status service`,
	Run: func(cmd *cobra.Command, args []string) {
		cc := exec.Command("systemctl", "status", tools.CurrentName()+".service")
		cc.Stdout = os.Stdout
		//异步启动子进程
		cc.Run()
	},
}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "log",
	Long:  `log service`,
	Run: func(cmd *cobra.Command, args []string) {
		fname := tools.CurrentName()
		var cc *exec.Cmd
		if len(args) > 0 && args[0] == "cat" {
			cc = exec.Command("cat", "status", "/var/log/"+fname+".log")
		} else {
			cc = exec.Command("tail", "-f", "/var/log/"+fname+".log")
		}
		cc.Stdout = os.Stdout
		//异步启动子进程
		cc.Run()
	},
}

func run(name string, arg ...string) error {
	daemoCmd := exec.Command(name, arg...)
	//异步启动子进程
	return daemoCmd.Start()
}

func ctlSvc(ctl string) error {
	err := run("systemctl", ctl, tools.CurrentName()+".service")
	if err == nil {
		log.Println(ctl + " service success")
	} else {
		log.Println(ctl+" service error", err)
	}
	return err
}

func init() {
	RootCmd.AddCommand(logCmd)
	RootCmd.AddCommand(svcCmd)
	RootCmd.AddCommand(installCmd)
	RootCmd.AddCommand(uninstallCmd)
	RootCmd.AddCommand(statusCmd)
	RootCmd.AddCommand(startCmd)
	RootCmd.AddCommand(stopCmd)
}

const svcText = `[Unit]
Description=%s
After=network.target
 
[Service]
Type=simple
WorkingDirectory=%s
ExecStart=%s/%s server log
Restart=on-failure
 
[Install]
WantedBy=multi-user.target`
