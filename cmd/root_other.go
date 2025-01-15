//go:build !windows
// +build !windows

package cmd

import (
	"fmt"
	"os"
	"sync"
	"syscall"
)

type LazyFile struct {
	file *os.File
	mu   sync.Mutex
}

func (lf *LazyFile) Write(p []byte) (n int, err error) {
	lf.mu.Lock()
	defer lf.mu.Unlock()
	if lf.file == nil {
		// 创建文件
		var err error
		lf.file, err = os.Create("stderr.log")
		if err != nil {
			return 0, err
		}
	}
	return lf.file.Write(p)
}

func redirectPanic() *os.File {
	file, err := os.OpenFile("panic.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法打开文件: %v\n", err)
		return nil
	} else if err = syscall.Dup2(int(file.Fd()), syscall.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "重定向错误输出失败: %v\n", err)
		file.Close()
		return nil
	}
	return file
}
