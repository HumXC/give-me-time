package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Info struct {
	Name        string `json:"name"`
	Discription string `json:"discription"`
	Version     string `json:"version"`
}

// 从 file 加载 json 文件，反序列化成 Info 并验证 Info 的正确性
// 内部已经调用了 VerifyInfo
func LoadInfo(file string) (*Info, error) {
	infoB, err := os.ReadFile(file)
	info := new(Info)
	makeErr := func(err error) error {
		return fmt.Errorf("failed to load info: %w", err)
	}
	if err != nil {
		return nil, makeErr(err)
	}
	err = json.Unmarshal(infoB, info)
	if err != nil {
		return nil, makeErr(err)
	}
	err = VerifyInfo(*info)
	if err != nil {
		return nil, makeErr(err)
	}
	return info, nil
}

// 检查 Info 中的内容是否符合要求：
// - Name 不能为空
func VerifyInfo(opt Info) error {
	if opt.Name == "" {
		return fmt.Errorf("field [name] is empty in info")
	}
	return nil
}
