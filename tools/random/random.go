package random

import (
	"math/rand"
	"strings"
	"time"
)

// Charsets
const (
	Uppercase  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Lowercase  = "abcdefghijklmnopqrstuvwxyz"
	Number     = "0123456789"
	Alphabetic = Uppercase + Lowercase + Number
)

func Int(min, max int) int {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rnd.Intn(max-min) + min
}

func Str(length uint8, charsets ...string) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := strings.Join(charsets, "")
	if charset == "" {
		charset = Alphabetic
	}
	b := make([]byte, length)
	count := int64(len(charset))
	for i := range b {
		b[i] = charset[rnd.Int63()%count]
	}
	return string(b)
}
