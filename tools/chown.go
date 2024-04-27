//go:build !linux
// +build !linux

package tools

import (
	"os"
)

func Chown(_ string, _ os.FileInfo) error {
	return nil
}
