package engine

import (
	"encoding/json"
	"fmt"
	"os"
)

type Option struct {
	Name        string    `json:"name"`
	Discription string    `json:"discription"`
	Version     string    `json:"version"`
	Element     []Element `json:"element"`
}

type Element struct {
	Name        string    `json:"name"`
	Discription string    `json:"discription"`
	Src         string    `json:"src"`  // 元素对应的图片
	Area        Area      `json:"area"` // 元素对应的区域，只有当没有 Src 时才会检查 Area
	Element     []Element `json:"element"`
	Offset      struct {  // 该元素在 Src 或 Area 上的偏移位置
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"offset"`
}

// 从左上角的点坐标到右下角的点坐标
type Area struct {
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
	X2 int `json:"x2"`
	Y2 int `json:"y2"`
}

// 从 file 加载 json 文件，反序列化成 Option 并验证 Option 的正确性
// 内部已经调用了 VerifyOption，VerifyElement
func LoadOption(file string) (*Option, error) {
	optB, err := os.ReadFile(file)
	opt := new(Option)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(optB, opt)
	if err != nil {
		return nil, err
	}
	return opt, VerifyOption(opt)
}

// 检查 Option 中的内容是否符合要求：
// - Name 不能为空
// 内部已经调用了 VerifyElement
func VerifyOption(opt *Option) error {
	if opt.Name == "" {
		return fmt.Errorf("option name is empty")
	}
	VerifyElement("", opt.Element)
	return nil
}

// 检查 Element 中的内容是否符合要求：
// - Name 不能为空
// - 同节点下 Name 不能重复
func VerifyElement(name string, es []Element) error {
	if len(es) == 0 {
		return nil
	}
	m := make(map[string]struct{})
	if name != "" {
		name += "."
	}
	for _, e := range es {
		if e.Name == "" {
			return fmt.Errorf("element name is empty in [%s]", name)
		}
		if _, ok := m[e.Name]; ok {
			return fmt.Errorf("element name can not be repeat in [%s]", name)
		}
		m[e.Name] = struct{}{}
		VerifyElement(name+e.Name, e.Element)
	}
	return nil
}