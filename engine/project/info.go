package project

import (
	"fmt"
	"os"
	"runtime"

	"gopkg.in/yaml.v3"
)

type Info struct {
	Name        string  `yaml:"name"`
	Discription string  `yaml:"discription"`
	Version     string  `yaml:"version"`
	Runtime     Runtime `yaml:"runtime"`
}

// 脚本代码的运行环境，Name 就是 Name，例如 go，nodejs，python
// Health 是用于检查运行环境状态的命令，如果返回码不为 0 则视为失败。
// Run 是执行代码的命令，例如 go run main.go .
// 带有操作系统后缀的 Health_* 和 Run_* 会根据对应的操作系统选择
// 正确的选项覆盖 Verify 和 Run
// BUG: 使用 json 的 tag 时，带有下划线和连字符的字段无法被反序列化
type Runtime struct {
	Name          string `yaml:"name"`
	Health        string `yaml:"health"`
	HealthWindows string `yaml:"health_windows"`
	HealthLinux   string `yaml:"health_linux"`
	Run           string `yaml:"run"`
	RunWindows    string `yaml:"run_windows"`
	RunLinux      string `yaml:"run_linux"`
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
	err = yaml.Unmarshal(infoB, info)
	if err != nil {
		return nil, makeErr(err)
	}

	// 根据当前的操作系统覆盖 Health 和 Run
	health := info.Runtime.Health
	run := info.Runtime.Run
	switch runtime.GOOS {
	case "linux":
		if info.Runtime.HealthLinux != "" {
			health = info.Runtime.HealthLinux
		}
		if info.Runtime.RunLinux != "" {
			run = info.Runtime.RunLinux
		}

	case "windows":
		if info.Runtime.HealthWindows != "" {
			health = info.Runtime.HealthWindows
		}
		if info.Runtime.RunWindows != "" {
			run = info.Runtime.RunWindows
		}
	}
	info.Runtime.Health = health
	info.Runtime.Run = run

	err = VerifyInfo(*info)
	if err != nil {
		return nil, makeErr(err)
	}
	return info, nil
}

// 检查 Info 中的内容是否符合要求：
// - Name, Runtime.Name, Runtime.Run 不能为空
func VerifyInfo(info Info) error {
	if info.Name == "" {
		return fmt.Errorf("field [name] cannot be empty in info")
	}
	if info.Runtime.Name == "" {
		return fmt.Errorf("field [runtime.name] cannot be empty in info")
	}
	if info.Runtime.Run == "" {
		return fmt.Errorf("field [runtime.run] cannot be empty in info")
	}
	return nil
}
