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

// 从 file 加载 json 文件，反序列化成 Option 并验证 Option 的正确性
// 内部已经调用了 VerifyOption
func LoadInfo(file string) (*Info, error) {
	optB, err := os.ReadFile(file)
	opt := new(Info)
	makeErr := func(err error) error {
		return fmt.Errorf("failed to load option: %w", err)
	}
	if err != nil {
		return nil, makeErr(err)
	}
	err = json.Unmarshal(optB, opt)
	if err != nil {
		return nil, makeErr(err)
	}
	err = VerifyOption(opt)
	if err != nil {
		return nil, makeErr(err)
	}
	return opt, nil
}

// 检查 Option 中的内容是否符合要求：
// - Name 不能为空
func VerifyOption(opt *Info) error {
	if opt.Name == "" {
		return fmt.Errorf("field [name] is empty in option")
	}
	return nil
}
