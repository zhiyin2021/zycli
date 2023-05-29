package tools

import (
	"os"
)

// 解析配置文件
func LoadConfig[T any](filename string, unmarshal func([]byte, any) error) (T, error) {
	var _config T
	buf, err := readCfg(filename)
	if err != nil {
		return _config, err
	}
	err = unmarshal(buf, &_config)
	return _config, err
}

func readCfg(filename string) ([]byte, error) {
	path := CurrentDir() + filename
	if FileExist(path) {
		return os.ReadFile(path)
	}
	return os.ReadFile(filename)
}
