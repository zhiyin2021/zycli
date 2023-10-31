package main

import (
	"net"

	"github.com/sirupsen/logrus"
	"github.com/zhiyin2021/zycli/cmd"
)

func main() {
	if cmd.IsRuning() {
		logrus.Errorln("already running")
		return
	}
	listen, err := net.Listen("unix", "test.ipc")
	if err != nil {
		logrus.Errorln("usock err:", err)
	}
	logrus.Println("usock success")
	for {
		conn, err := listen.Accept()
		if err != nil {
			logrus.Errorln("usock err:", err)
			continue
		}
		for {
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				logrus.Errorln("read err:", err)
				break
			}
			logrus.Println("read:", string(buf[:n]))
		}
	}
}
