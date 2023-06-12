package tools

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net"
	"os"
	"os/exec"
	"reflect"
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

// QuickSort 快速排序,key不为空按结构体识别排序,否则按数字类型排序
func QuickSort[T any](arr []T, key string) []T {
	if len(arr) <= 1 {
		return arr
	}
	low, high := 0, len(arr)-1
	if key != "" {
		f1 := reflect.ValueOf(arr[0])
		v0 := f1.FieldByName(key).Int()
		for i := 1; i <= high; {
			if reflect.ValueOf(arr[i]).FieldByName(key).Int() > v0 {
				arr[i], arr[high] = arr[high], arr[i]
				high--
			} else {
				arr[i], arr[low] = arr[low], arr[i]
				low++
				i++
			}
		}
	} else {
		v0 := toi64(arr[0])
		for i := 1; i <= high; {
			if toi64(arr[i]) > v0 {
				arr[i], arr[high] = arr[high], arr[i]
				high--
			} else {
				arr[i], arr[low] = arr[low], arr[i]
				low++
				i++
			}
		}
	}
	QuickSort(arr[:low], key)
	QuickSort(arr[low+1:], key)
	return arr
}

func toi64(n interface{}) int64 {
	switch n := n.(type) {
	case int:
		return int64(n)
	case int8:
		return int64(n)
	case int16:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	}
	return 0
}
