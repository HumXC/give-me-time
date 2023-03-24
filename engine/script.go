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

type script struct {
	l    *lua.State
	file string
}

func (s script) Run() error {
	return lua.DoFile(s.l, s.file)
}

func (s script) File() string {
	return s.file
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
	L := lua.NewState()
	lua.OpenLibraries(L)
	setElement(L, option.Element)
	setFunction(L, api)
	return script{
		l:    L,
		file: file,
	}
}

// 设置在 lua 中的全局 E
func setElement(l *lua.State, es []Element) {
	l.NewTable()
	defer func() {
		l.SetGlobal("E")
	}()
	pushElement(l, "", es)
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

// 设置在 lua 中的全局函数
func setFunction(L *lua.State, api Api) {

	L.Register("press", func(l *lua.State) (rt int) {

		if name := parseElement(l, 1); name != "" {
			duration, err := getDuration(l, 2)
			if err != nil {
				pushErr(l, err)
				return
			}
			e := getElement(name)
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
	L.Register("swipe", func(l *lua.State) (rt int) {
		rt = 1
		var sh SwipeHandler
		if name := parseElement(l, 1); name != "" {
			e := getElement(name)
			sh = api.SwipeE(e)
			setSwipeHandler(l, sh)
			return
		}
		x, y, err := getXY(l, 1, 2)
		if err != nil {
			pushErr(l, err)
			return
		}
		sh = api.Swipe(x, y)
		setSwipeHandler(l, sh)
		return
	})

}
func pushErr(l *lua.State, err error) {
	lua.Errorf(l, err.Error())
}
func getElement(name string) Element {
	// TODO: implement 根据 name 获取元素，name 是形如 "main.start" 的字符串
	return Element{
		Discription: name,
	}
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

func setSwipeHandler(l *lua.State, sh SwipeHandler) {
	if sh == nil {
		return
	}
	l.NewTable()

	l.PushString("to")
	l.PushGoFunction(func(state *lua.State) (rt int) {
		rt = 1
		if name := parseElement(l, 1); name != "" {
			e := getElement(name)
			sh.ToE(e)
			setSwipeHandler(l, sh)
			return
		}
		x, y, err := getXY(l, 1, 2)
		if err != nil {
			pushErr(l, err)
			return
		}
		sh.To(x, y)
		setSwipeHandler(l, sh)
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
