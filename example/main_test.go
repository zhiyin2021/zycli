package main

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/zhiyin2021/zycli/tools"
	"github.com/zhiyin2021/zycli/tools/cache"
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

func TestCache(t *testing.T) {
	c := cache.NewMemory(context.Background())
	c.SetBySliding("a", "b", time.Second*5)
	for i := 0; i < 11; i++ {
		time.Sleep(time.Second)
		if i == 8 {
			c.Del("a")
		}
		if i == 9 {
			time.Sleep(time.Second * 6)
		}
		log.Println("get1=>", c.Get("a"))
	}

	for i := 0; i < 15; i++ {
		time.Sleep(time.Second)
		c.Del("a")
		log.Println("get2=>", c.Get("a"))
	}
}
