package main

import (
	"net"

	"github.com/sirupsen/logrus"
)

func main() {
	dial, err := net.Dial("unix", "test.sock")
	if err != nil {
		logrus.Errorln("dial err:", err)
	} else {
		logrus.Println("dial success")
		defer dial.Close()
	}
}
