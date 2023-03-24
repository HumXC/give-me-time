package engine

import (
	"fmt"

	"github.com/Shopify/go-lua"
)

type Script interface {
	// 开始运行脚本
	Run() error
	// 返回文件名
	File() string
}
type Storage struct {
	element map[string]Element
}

func (s *Storage) Element(key string) Element {
	return s.element[key]
}

type script struct {
	l       *lua.State
	file    string
	storage Storage
}

func (s *script) Run() error {
	return lua.DoFile(s.l, s.file)
}

func (s *script) File() string {
	return s.file
}

// 设置在 lua 中的全局函数
func (s *script) setFunction(api Api) {
	s.l.Register("press", func(l *lua.State) (rt int) {
		if name := parseElement(l, 1); name != "" {
			duration, err := getDuration(l, 2)
			if err != nil {
				pushErr(l, err)
				return
			}
			e := s.storage.Element(name)
			err = api.PressE(e, duration)
			if err != nil {
				pushErr(l, err)
				return
			}
			return
		}
		duration, err := getDuration(l, 3)
		if err != nil {
			pushErr(l, err)
			return
		}
		x, y, err := getXY(l, 1, 2)
		if err != nil {
			pushErr(l, err)
			return
		}
		err = api.Press(x, y, duration)
		if err != nil {
			pushErr(l, err)
			return
		}
		return
	})
	s.l.Register("swipe", func(l *lua.State) (rt int) {
		rt = 1
		var sh SwipeHandler
		if name := parseElement(l, 1); name != "" {
			e := s.storage.Element(name)
			sh = api.SwipeE(e)
			s.setSwipeHandler(l, sh)
			return
		}
		x, y, err := getXY(l, 1, 2)
		if err != nil {
			pushErr(l, err)
			return
		}
		sh = api.Swipe(x, y)
		s.setSwipeHandler(l, sh)
		return
	})

}

func (s *script) setSwipeHandler(l *lua.State, sh SwipeHandler) {
	if sh == nil {
		return
	}
	l.NewTable()

	l.PushString("to")
	l.PushGoFunction(func(state *lua.State) (rt int) {
		rt = 1
		if name := parseElement(l, 1); name != "" {
			e := s.storage.Element(name)
			sh.ToE(e)
			s.setSwipeHandler(l, sh)
			return
		}
		x, y, err := getXY(l, 1, 2)
		if err != nil {
			pushErr(l, err)
			return
		}
		sh.To(x, y)
		s.setSwipeHandler(l, sh)
		return
	})
	l.SetTable(-3)

	l.PushString("action")
	l.PushGoFunction(func(state *lua.State) (rt int) {
		rt = 1
		duration, err := getDuration(l, 1)
		if err != nil {
			pushErr(l, err)
			return
		}
		err = sh.Action(duration)
		if err != nil {
			pushErr(l, err)
			return
		}
		return
	})
	l.SetTable(-3)
}

// 设置在 lua 中的全局 E
func (s *script) setElement(es []Element) {
	s.l.NewTable()
	defer func() {
		s.l.SetGlobal("E")
	}()
	pushElement(s.l, "", es)
}

// “Element” 类型参数是指在 lua 中以 “click(E.main.start)” 的形式调用
type Api interface {
	// 按下一个元素或坐标，duration 单位是 ms。duration 为 0 时，将会自动把 duration 赋值为 100
	Press(x, y, duration int) error
	PressE(e Element, duration int) error
	// 滑动
	Swipe(x, y int) SwipeHandler
	SwipeE(Element) SwipeHandler
}

// Api 中的 Swipe 函数返回给 lua 一个 SwipeHandler
// SwipeHandler 是可以链式调用的
type SwipeHandler interface {
	To(x, y int) SwipeHandler
	ToE(Element) SwipeHandler
	Action(duration int) error
}

func LoadScript(file string, option *Option, api Api) Script {
	s := &script{
		l:    lua.NewState(),
		file: file,
		storage: Storage{
			element: make(map[string]Element),
		},
	}
	lua.OpenLibraries(s.l)
	StoreElement(s.storage.element, "", option.Element)
	s.setElement(option.Element)
	s.setFunction(api)

	return s
}

// 扁平化 Element 存储到 map 中，Element.Element 将被赋值为 nil 不再嵌套
func StoreElement(m map[string]Element, name string, es []Element) {
	if len(es) == 0 {
		return
	}
	if name != "" {
		name += "."
	}
	for _, e := range es {
		subE := e.Element
		_name := name + e.Name
		e.Element = nil
		m[_name] = e
		StoreElement(m, _name, subE)
	}
}

func pushElement(l *lua.State, name string, es []Element) {
	if len(es) == 0 {
		return
	}
	if name != "" {
		name += "."
	}
	push := func(k, v string) {
		l.PushString(k)
		l.PushString(v)
		l.SetTable(-3)
	}
	for _, e := range es {
		_name := name + e.Name
		l.PushString(e.Name)
		l.NewTable()

		push("_name", _name)
		pushElement(l, _name, e.Element)

		l.SetTable(-3)
	}
}

func pushErr(l *lua.State, err error) {
	lua.Errorf(l, err.Error())
}
func getDuration(l *lua.State, index int) (int, error) {
	v := l.ToValue(index)
	if v == nil {
		return 100, nil
	}
	d, ok := l.ToInteger(index)
	if !ok {
		return 0, fmt.Errorf("duration [%v] is not an integer", v)
	}
	if d < 0 {
		return 0, fmt.Errorf("duration [%d] is not an positive integer", d)
	}
	if d == 0 {
		d = 100
	}
	return d, nil
}
func getXY(l *lua.State, indexX, indexY int) (int, int, error) {
	x, ok := l.ToInteger(indexX)
	if !ok {
		err := fmt.Errorf("x [%v] is not an integer", l.ToValue(1))
		return 0, 0, err
	}
	y, ok := l.ToInteger(indexY)
	if !ok {
		err := fmt.Errorf("y [%v] is not an integer", l.ToValue(2))
		return 0, 0, err
	}
	return x, y, nil
}

// 用在 lua.Function 中，判段第 index 个参数是不是 Element
// 如果是，则返回 Element 的 "_name"
func parseElement(L *lua.State, index int) string {
	if L.TypeOf(index) != lua.TypeTable {
		return ""
	}
	L.Field(index, "_name")
	s, ok := L.ToString(-1)
	if !ok {
		return ""
	}
	return s
}
