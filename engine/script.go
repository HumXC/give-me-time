package engine

import (
	"fmt"
	"time"

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
	err := lua.DoFile(s.l, s.file)
	if err != nil {
		return fmt.Errorf("script run error: %w", err)
	}
	return nil
}

func (s *script) File() string {
	return s.file
}

// 设置在 lua 中的全局函数
func (s *script) setFunction(api Api) {
	s.l.Register("sleep", func(l *lua.State) (rt int) {
		duration, err := getDuration(l, 1)
		if err != nil {
			pushErr(l, err)
			return
		}
		time.Sleep(time.Duration(duration) * time.Millisecond)
		return
	})
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
		setSwipeAction := func(l *lua.State, st SwipeAction) (rt int) {
			rt = 1
			l.NewTable()
			l.PushString("action")
			l.PushGoFunction(func(state *lua.State) (rt int) {
				rt = 1
				duration, err := getDuration(l, 1)
				if err != nil {
					pushErr(l, err)
					return
				}
				err = st.Action(duration)
				if err != nil {
					pushErr(l, err)
					return
				}
				return
			})
			l.SetTable(-3)
			return
		}
		setSwipeTo := func(l *lua.State, st SwipeTo) (rt int) {
			rt = 1
			l.NewTable()
			l.PushString("to")
			l.PushGoFunction(func(state *lua.State) (rt int) {
				rt = 1
				if name := parseElement(l, 1); name != "" {
					e := s.storage.Element(name)
					sac := st.ToE(e)
					setSwipeAction(l, sac)
					return
				}
				x, y, err := getXY(l, 1, 2)
				if err != nil {
					pushErr(l, err)
					return
				}
				sac := st.To(x, y)
				setSwipeAction(l, sac)
				return
			})
			l.SetTable(-3)
			return
		}

		if name := parseElement(l, 1); name != "" {
			e := s.storage.Element(name)
			st := api.SwipeE(e)
			setSwipeTo(l, st)
			return
		}
		x, y, err := getXY(l, 1, 2)
		if err != nil {
			pushErr(l, err)
			return
		}
		st := api.Swipe(x, y)
		setSwipeTo(l, st)
		return
	})

}

// 设置在 lua 中的全局 E
func (s *script) setElement(es []Element) {
	s.l.NewTable()
	defer func() {
		s.l.SetGlobal("E")
	}()
	pushElement(s.l, "", es)
}

func LoadScript(file string, option *Option, element []Element, api Api) Script {
	s := &script{
		l:    lua.NewState(),
		file: file,
		storage: Storage{
			element: make(map[string]Element),
		},
	}
	lua.OpenLibraries(s.l)
	StoreElement(s.storage.element, "", element)
	s.setElement(element)
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
