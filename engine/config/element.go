package config

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path"
)

const (
	ElTypeImg   = "img"
	ElTypeArea  = "area"
	ElTypePoint = "point"
)

// Element 一般用于图像识别
// 其中 Img, Area, Point 三者起到的作用相同，用于表达一片区域(Img, Area)或者一个点(Point),
// 这 3 者只能有其一发挥作用，优先级为：Img > Area > Point,
// 也就是说当 Img 不为空时，Area 和 Point 的值不会有作用。
// Offset 是相对 Img 或者 Area 的偏移量
type Element struct {
	Type        string
	Name        string      `json:"name"`
	Discription string      `json:"discription"`
	Img         string      `json:"img"`
	Area        Area        `json:"area"`
	Point       image.Point `json:"point"`
	Element     []Element   `json:"element"`
	Offset      image.Point `json:"offset"` // 该元素在 Img 或 Area 上的偏移位置
	Threshold   float32     `json:"threshold"`
}
type ElImg struct {
	Discription string
	Img         []byte
	Offset      image.Point
	Threshold   float32
}
type ElArea struct {
	Discription string
	P1, P2      image.Point
}
type ElPoint struct {
	image.Point
	Discription string
}

// 从左上角的点坐标到右下角的点坐标
type Area struct {
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
	X2 int `json:"x2"`
	Y2 int `json:"y2"`
}

// 从 file 加载 json 文件，反序列化成 Element 并验证 Element 的正确性
// 内部已经调用了 VerifyElement
func LoadElement(file string) ([]Element, error) {
	e := make([]Element, 0)
	makeErr := func(err error) error {
		return fmt.Errorf("failed to load element: %w", err)
	}
	eB, err := os.ReadFile(file)
	if err != nil {
		return nil, makeErr(err)
	}
	m := make([]map[string]any, 0)
	err = json.Unmarshal(eB, &e)
	if err != nil {
		return nil, makeErr(err)
	}
	err = json.Unmarshal(eB, &m)
	if err != nil {
		return nil, makeErr(err)
	}
	// 赋值 Type
	e = SetType(e, m)
	err = VerifyElement("", e)
	if err != nil {
		return nil, makeErr(err)
	}
	PatchAbsPath(e, path.Dir(file))
	return e, nil
}

func SetType(es []Element, ms []map[string]any) []Element {
	if len(es) == 0 {
		return nil
	}
	for i := 0; i < len(es); i++ {
		switch {
		case ms[i][ElTypeImg] != nil:
			es[i].Type = ElTypeImg
		case ms[i][ElTypeArea] != nil:
			es[i].Type = ElTypeArea
		case ms[i][ElTypePoint] != nil:
			es[i].Type = ElTypePoint
		}
		if es[i].Element != nil {
			// ms[i]["element"] 无法直接断言成 []map[string]any
			m := make([]map[string]any, 0)
			_ms, _ := ms[i]["element"].([]any)
			for _, _m := range _ms {
				v, _ := _m.(map[string]any)
				m = append(m, v)
			}
			es[i].Element = SetType(es[i].Element, m)
		}
	}
	return es
}

// 检查 Element 中的内容是否符合要求：
// - Name 不能为空
// - 同节点下 Name 不能重复
// - 如果 Type 不为空，则 Type 必须是已经定义的
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
			return fmt.Errorf("element name [%s] can not be repeat in [%s]", e.Name, name)
		}
		switch e.Type {
		case "":
		case ElTypeImg:
		case ElTypeArea:
		case ElTypePoint:
		default:
			return fmt.Errorf("element [%s] type [%s] undefined %v ", e.Name, name,
				[]string{ElTypeImg, ElTypeArea, ElTypePoint})
		}
		m[e.Name] = struct{}{}
		VerifyElement(name+e.Name, e.Element)
	}
	return nil
}

func ParseElement(elements []Element) (map[string]ElImg, map[string]ElArea, map[string]ElPoint, error) {
	elImg := make(map[string]ElImg)
	elArea := make(map[string]ElArea)
	elPoint := make(map[string]ElPoint)
	fElement := make(map[string]Element)
	storeImg := func(k string, e Element) error {
		b, err := os.ReadFile(e.Img)
		elImg[k] = ElImg{
			Discription: e.Discription,
			Img:         b,
			Offset:      e.Offset,
			Threshold:   e.Threshold,
		}
		return err

	}
	storeArea := func(k string, e Element) {
		elArea[k] = ElArea{
			Discription: e.Discription,
			P1:          image.Pt(e.Area.X1, e.Area.Y1),
			P2:          image.Pt(e.Area.X2, e.Area.Y2),
		}
	}
	storePoint := func(k string, e Element) {
		elPoint[k] = ElPoint{
			Discription: e.Discription,
			Point:       image.Pt(e.Point.X, e.Point.Y),
		}
	}
	FlatElement(fElement, "", elements)
	for k, e := range fElement {
		switch e.Type {
		case ElTypeImg:
			err := storeImg(k, e)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to parse element [%s]: %w", k, err)
			}
		case ElTypeArea:
			storeArea(k, e)
		case ElTypePoint:
			storePoint(k, e)
		}
	}
	return elImg, elArea, elPoint, nil
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
