package tools

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/zhiyin2021/zycli/tools/logger"
)

func If[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

func FileExists(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func GenId() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// RunCmd 执行命令,返回内容
func RunCmd(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	res := out.String()
	if err != nil {
		logger.Errorln("RunCmd", args, err)
	} else {
		logger.Infoln("RunCmd", args, "=>", res)
	}
	return res
}
func GetIpList() []string {
	ips := []string{}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Errorln("getIp ", err)
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

func AtoI(s string) int {
	s = strings.TrimSpace(s)
	n, e := strconv.Atoi(s)
	if e != nil {
		return 0
	}
	return n
}
func AtoU64(s string) uint64 {
	s = strings.TrimSpace(s)
	n, e := strconv.ParseUint(s, 0, 0)
	if e != nil {
		return 0
	}
	return n
}
func AtoI64(s string) int64 {
	n, e := strconv.ParseInt(s, 0, 0)
	if e != nil {
		return 0
	}
	return n
}
func AtoF(s string) float64 {
	n, e := strconv.ParseFloat(s, 64)
	if e != nil {
		return 0
	}
	return n
}
func ToStr(s any) string {
	return fmt.Sprintf("%v", s)
}

func IsEmpty(content string) bool {
	return strings.TrimSpace(content) == ""
}

// 判断切片中是否有字符串
func SliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func JsonStr(data any) string {
	b, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(b)
}
func JsonMap(body []byte) map[string]any {
	var data map[string]any
	err := json.Unmarshal(body, &data)
	if err != nil {
		return nil
	}
	return data
}
func As[T any](v any) (o T) {
	if v1, ok := v.(T); ok {
		return v1
	}
	return o
}

func MapItem[T any](m map[string]string, key string, def T) (ret T) {
	ret = def
	if val, ok := m[key]; ok {
		v := reflect.ValueOf(&ret).Elem()
		switch v.Kind() {
		case reflect.String:
			v.SetString(val)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if n, err := strconv.ParseInt(val, 0, 0); err == nil {
				v.SetInt(n)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if n, err := strconv.ParseUint(val, 0, 0); err == nil {
				v.SetUint(n)
			}
		}
	}
	return
}
func RemoveElement(data []int32, target int32) ([]int32, bool) {
	index := -1
	// 查找目标值的索引
	for i, v := range data {
		if v == target {
			index = i
			break
		}
	}

	// 如果找到，删除该元素
	if index != -1 {
		data = append(data[:index], data[index+1:]...)
		return data, true
	}
	return data, false // 没有找到目标值
}
