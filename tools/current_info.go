package tools

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	pathOnce sync.Once
	nameOnce sync.Once
	curPath  string
	curName  string
)

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
		return "~/Library/Logs/"
	}
	return "/var/log/"
}
