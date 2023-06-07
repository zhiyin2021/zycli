package tools

import (
	"fmt"
	"sync"
)

type ProgressBar struct {
	percent int64  //百分比
	cur     int64  //当前进度位置
	total   int64  //总进度
	rate    string //进度条
	graph   string //显示符号
	mutex   sync.Mutex
}

// func main() {
// 	var bar Bar
// 	bar.NewOption(0, 100)
// 	//bar.NewOptionWithGraph(0, 100, "#")
// 	for i := 0; i <= 100; i++ {
// 		time.Sleep(100 * time.Millisecond)
// 		bar.Play(int64(i))
// 	}
// 	bar.Finish()
// }

func (bar *ProgressBar) NewOption(start, total int64) {
	bar.mutex.Lock()
	defer bar.mutex.Unlock()
	bar.cur = start
	bar.total = total
	if bar.graph == "" {
		bar.graph = "█"
	}
	bar.percent = bar.getPercent()
	for i := 0; i < int(bar.percent); i += 2 {
		bar.rate += bar.graph //初始化进度条位置
	}
}

func (bar *ProgressBar) getPercent() int64 {
	return int64(float32(bar.cur) / float32(bar.total) * 100)
}

func (bar *ProgressBar) NewOptionWithGraph(start, total int64, graph string) {
	bar.mutex.Lock()
	defer bar.mutex.Unlock()
	bar.graph = graph
	bar.NewOption(start, total)
}

func (bar *ProgressBar) Play(cur int64) {
	bar.mutex.Lock()
	defer bar.mutex.Unlock()
	bar.cur = cur
	last := bar.percent
	bar.percent = bar.getPercent()
	if bar.percent != last && bar.percent%2 == 0 {
		bar.rate += bar.graph
	}
	fmt.Printf("\r[%-50s]%3d%%  %8d/%d", bar.rate, bar.percent, bar.cur, bar.total)
}

func (bar *ProgressBar) Inc() {
	bar.mutex.Lock()
	defer bar.mutex.Unlock()
	bar.cur++
	last := bar.percent
	bar.percent = bar.getPercent()
	if bar.percent != last && bar.percent%2 == 0 {
		bar.rate += bar.graph
	}
	fmt.Printf("\r[%-50s]%3d%%  %8d/%d", bar.rate, bar.percent, bar.cur, bar.total)
}
func (bar *ProgressBar) Printf(msg string, a ...interface{}) {
	s := fmt.Sprintf(msg, a...)
	fmt.Printf("\r%-100s", s)
}
func (bar *ProgressBar) Finish() {
	fmt.Println()
}
