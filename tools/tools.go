package tools

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

func If[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

func FileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}

func GenId() string {
	id, _ := uuid.NewV4()
	return strings.ReplaceAll(id.String(), "-", "")
}

// RunCmd 执行命令,返回内容
func RunCmd(name string, args ...string) string {
	// logrus.Info("cmd", "runCmd:", name)
	// arg := append([]string{"-c"}, args...)
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	res := out.String()
	if err != nil {
		logrus.Errorln("RunCmd", args, err)
	} else {
		logrus.Infoln("RunCmd", args, "=>", res)
	}
	return res
}
func GetIpList() []string {
	ips := []string{}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logrus.Error("getIp ", err)
	} else {
		for _, address := range addrs {
			// 检查ip地址判断是否回环地址
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ips = append(ips, ipnet.IP.String())
				}
			}
		}
	}
	ips = append(ips, "127.0.0.1")
	return ips
}

func Md5(data string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}
