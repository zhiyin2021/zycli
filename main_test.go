package main

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/zhiyin2021/zycli/resp"
	"github.com/zhiyin2021/zycli/tools"
)

func EchoTest(t *testing.T) {
	e := resp.GetEcho()
	e.GET("/api/", func(c resp.Context) error {
		return c.String(200, "ok")
	})
}

func RunlastTest(t *testing.T) {
	r := tools.NewRunLast(500 * time.Millisecond)
	var wg sync.WaitGroup
	tm := time.Now()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			r.Run(func() {
				log.Println("run last", idx)
			})
		}(i)
		time.Sleep(time.Millisecond * 20)
	}
	wg.Wait()
	log.Println("cost", time.Since(tm))
}
