package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Option struct {
	Name         string `json:"name"`
	Type         string `json:"type"` // string, number, bool
	Discription  string `json:"discription"`
	Select       []any  `json:"select"`
	IsOnlySelect bool   `json:"is_only_select"`
	Default      any    `json:"default"`
}

func LoadOption(file string) ([]Option, error) {
	optB, err := os.ReadFile(file)
	opts := make([]Option, 0)
	makeErr := func(err error) error {
		return fmt.Errorf("failed to load option: %w", err)
	}
	if err != nil {
		return nil, makeErr(err)
	}
	err = json.Unmarshal(optB, &opts)
	if err != nil {
		return nil, makeErr(err)
	}
	err = VerifyOption(opts)
	// 给 Default 赋予初值
	for i := 0; i < len(opts); i++ {
		var defaul any
		switch opts[i].Type {
		case "string":
			defaul = ""
		case "number":
			defaul = 0.0
		case "bool":
			defaul = false
		}
		if opts[i].Default == nil {
			opts[i].Default = defaul
		}
	}
	if err != nil {
		return nil, makeErr(err)
	}
	return opts, nil
}

// 检查 Option 中的内容是否符合要求：
//   - Name 和 Type 不能为空
//   - Name 不可重复
//   - Default 和 Select 的值要符合 Type 的定义，
//     如果 Type 为 string，那么 Default 就应该是 string，Select 就应该是 []string
func VerifyOption(opts []Option) error {
	assert := func(v any, t string) bool {
		var ok bool
		switch t {
		case "string":
			_, ok = v.(string)
		case "number":
			_, ok = v.(float64)
		case "bool":
			_, ok = v.(bool)
		}
		return ok
	}
	keys := make(map[string]struct{})
	for _, opt := range opts {
		if _, ok := keys[opt.Name]; ok {
			return fmt.Errorf("[name:%s] is already used", opt.Name)
		}
		if opt.Name == "" {
			return errors.New("[name] cannot be empty")
		}
		t := opt.Type
		if t == "" {
			return errors.New("[type] cannot be empty")
		}
		if !(t == "number" || t == "string" || t == "bool") {
			return errors.New("[type] must be [string|number|bool]")
		}
		if !assert(opt.Default, opt.Type) {
			return fmt.Errorf("[default:%v] does not match the [type:%s]", opt.Default, opt.Type)
		}
		for _, s := range opt.Select {
			if !assert(s, opt.Type) {
				return fmt.Errorf("[select:%v] does not match the [type:%s]", s, opt.Type)
			}
		}
		keys[opt.Name] = struct{}{}
	}
	return nil
}

func ParseOption(opts []Option, userOption string) (map[string]any, error) {
	optB, err := os.ReadFile(userOption)
	makeErr := func(err error) error {
		return fmt.Errorf("failed to parse option: %w", err)
	}
	assert := func(v any, t string) bool {
		var ok bool
		switch t {
		case "string":
			_, ok = v.(string)
		case "number":
			_, ok = v.(float64)
		case "bool":
			_, ok = v.(bool)
		}
		return ok
	}
	if err != nil {
		return nil, makeErr(err)
	}
	m := make(map[string]any, 0)
	err = json.Unmarshal(optB, &m)
	if err != nil {
		return nil, makeErr(err)
	}
	result := make(map[string]any)
	for _, opt := range opts {
		v, ok := m[opt.Name]
		if !ok {
			result[opt.Name] = opt.Default
			continue
		}
		if assert(v, opt.Type) {
			result[opt.Name] = v
			continue
		} else {
			return nil, fmt.Errorf("[name:%s] name uses the [%s] type", opt.Name, opt.Type)
		}
	}
	return result, nil
}
