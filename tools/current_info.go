package tools

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
	pathOnce sync.Once
	nameOnce sync.Once
	curPath  string
	curName  string
)

// 获取当前执行程序所在的绝对路径
func FixPath(f string) string {
	if !strings.Contains(f, "/") {
		return CurrentDir() + "/" + f
	}
	return f
}

// 获取当前执行程序所在的绝对路径
func CurrentDir() string {
	pathOnce.Do(func() {
		exePath, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}
		curPath, _ = filepath.EvalSymlinks(filepath.Dir(exePath))
	})
	return curPath
}

// 获取当前执行程序所在的绝对路径
func CurrentName() string {
	nameOnce.Do(func() {
		path, _ := os.Executable()
		_, curName = filepath.Split(path)
	})
	return curName
}

// 获取当前执行程序所在的绝对路径
func LogPath() string {
	if runtime.GOOS == "darwin" {
		return "./logs/"
	}
	return "/var/log/"
}
