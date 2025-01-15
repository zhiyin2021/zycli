package cmd

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func redirectPanic() *os.File {
	file, err := os.OpenFile("panic.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法打开文件: %v\n", err)
		return nil
	} else if err = windows.SetStdHandle(uint32(windows.Stderr), windows.Handle(file.Fd())); err != nil {
		fmt.Fprintf(os.Stderr, "重定向错误输出失败: %v\n", err)
		file.Close()
		return nil
	}
	return file
}
