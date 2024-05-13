package main

import (
	"fmt"
	"log"

	"github.com/zhiyin2021/zycli/cmd"
)

func main() {
	fmt.Println("start...")
	split := cmd.NewSplit("./log/test.log", cmd.OptMaxSize(100), cmd.OptMaxAge(7), cmd.OptCompressType(cmd.CT_XZ))
	log.SetOutput(split)
	fmt.Println("started")
	for {
		log.Println("hello world!!!hello world!!!hello world!!!hello world!!!hello world!!!hello world!!!hello world!!!")
		// time.Sleep(time.Second)
	}
}
