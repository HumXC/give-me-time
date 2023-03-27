package engine

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path"
	"strings"
)

type Option struct {
	Name        string `json:"name"`
	Discription string `json:"discription"`
	App         string `json:"app"`
	Version     string `json:"version"`
}

// Element 一般用于图像识别
// 其中 Img, Area, Point 三者起到的作用相同，用于表达一片区域(Img, Area)或者一个点(Point),
// 这 3 者只能有其一发挥作用，优先级为：Img > Area > Point,
// 也就是说当 Img 不为空时，Area 和 Point 的值不会有作用。
// Offset 是相对 Img 或者 Area 的偏移量
type Element struct {
	Name        string `json:"name"`
	Path        string
	Discription string      `json:"discription"`
	Img         string      `json:"img"`
	Area        Area        `json:"area"`
	Point       image.Point `json:"point"`
	Element     []Element   `json:"element"`
	Offset      image.Point `json:"offset"` // 该元素在 Img 或 Area 上的偏移位置
}

// 从左上角的点坐标到右下角的点坐标
type Area struct {
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
	X2 int `json:"x2"`
	Y2 int `json:"y2"`
}

// 从 file 加载 json 文件，反序列化成 Option 并验证 Option 的正确性
// 内部已经调用了 VerifyOption
func LoadOption(file string) (*Option, error) {
	optB, err := os.ReadFile(file)
	opt := new(Option)
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

// 从 file 加载 json 文件，反序列化成 Element 并验证 Element 的正确性
// 内部已经调用了 VerifyElement
func LoadElement(file string) ([]Element, error) {
	type E struct {
		Element []Element `json:"element"`
	}
	e := E{Element: make([]Element, 0)}
	makeErr := func(err error) error {
		return fmt.Errorf("failed to load element: %w", err)
	}
	esB, err := os.ReadFile(file)
	if err != nil {
		return nil, makeErr(err)
	}
	err = json.Unmarshal(esB, &e)
	if err != nil {
		return nil, makeErr(err)
	}
	err = VerifyElement("", e.Element)
	if err != nil {
		return nil, makeErr(err)
	}
	PatchAbsPath(e.Element, path.Dir(file))
	return e.Element, nil
}

// 检查 Option 中的内容是否符合要求：
// - Name 不能为空
func VerifyOption(opt *Option) error {
	if opt.Name == "" {
		return fmt.Errorf("field [name] is empty in option")
	}
	if opt.App == "" {
		return fmt.Errorf("field [app] is empty in option")
	}
	return nil
}

// 检查 Element 中的内容是否符合要求：
// - Name 不能为空
// - 同节点下 Name 不能重复
// - Name 不能含有字符 '.'
// - Name 不能包含 '-'
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
		if strings.Index(e.Name, ".") != -1 {
			return fmt.Errorf("element name [%s] is invalid in [%s]", e.Name, name)
		}
		if strings.Index(e.Name, "-") != -1 {
			return fmt.Errorf("element name [%s] is invalid in [%s]", e.Name, name)
		}
		if _, ok := m[e.Name]; ok {
			return fmt.Errorf("element name [%s] can not be repeat in [%s]", e.Name, name)
		}
		m[e.Name] = struct{}{}
		VerifyElement(name+e.Name, e.Element)
	}
	return nil
}

// 扁平化 Element 存储到 map 中，Element.Element 将被赋值为 nil 不再嵌套
// 并为 Element 的 Path 赋值
func FlatElement(m map[string]Element, name string, es []Element) {
	if len(es) == 0 {
		return
	}
	if name != "" {
		name += "."
	}
	for _, e := range es {
		subE := e.Element
		path := name + e.Name
		e.Element = nil
		e.Path = path
		m[path] = e
		FlatElement(m, path, subE)
	}
}

// 判断路径类型，修正相对路径
func PatchAbsPath(es []Element, basePath string) {
	if len(es) == 0 {
		return
	}
	for i := 0; i < len(es); i++ {
		if es[i].Img == "" {
			continue
		}
		if path.IsAbs(es[i].Img) {
			continue
		}
		es[i].Img = path.Join(basePath, es[i].Img)
		PatchAbsPath(es[i].Element, basePath)
	}
}
