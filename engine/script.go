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
	// // 滑动
	// Swipe(Element) (SwipeHandler, error)
	// SwipeE(x, y int) (SwipeHandler, error)
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
	L.Register("press", func(state *lua.State) (rt int) {
		getDuration := func(index int) (int, error) {
			v := L.ToValue(index)
			if v == nil {
				return 100, nil
			}
			d, ok := L.ToInteger(index)
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
		pushErr := func(err error) {
			lua.Errorf(L, err.Error())
		}
		if name := parseElement(L, 1); name != "" {
			duration, err := getDuration(2)
			if err != nil {
				pushErr(err)
				return
			}
			// TODO: implement
			err = api.PressE(Element{
				Discription: name,
			}, duration)
			if err != nil {
				pushErr(err)
				return
			}
			return
		}
		duration, err := getDuration(3)
		if err != nil {
			pushErr(err)
			return
		}
		x, ok := L.ToInteger(1)
		if !ok {
			err = fmt.Errorf("x [%v] is not an integer", L.ToValue(1))
			pushErr(err)
		}
		y, ok2 := L.ToInteger(2)
		if !ok || !ok2 {
			err := fmt.Errorf("y [%v] is not an integer", L.ToValue(2))
			pushErr(err)
			return
		}
		err = api.Press(x, y, duration)
		if err != nil {
			pushErr(err)
			return
		}
		return
	})
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
