package main

import (
	"log"
	"testing"

	"github.com/zhiyin2021/zycli/tools"
)

func TestQuickSort(t *testing.T) {
	cfg := []int32{32, 123, 589, 1, 3, 89, 12, 998, 234}
	log.Println(cfg)
	arr := tools.QuickSort(cfg, "")
	log.Println(arr)
}

func TestQuickSortStruct(t *testing.T) {
	cfg := []Config{
		{Port: 8080},
		{Port: 8083},
		{Port: 80889},
		{Port: 802},
		{Port: 850},
	}
	log.Println(cfg)
	arr := tools.QuickSort(cfg, "Port")
	log.Println(arr)
}
