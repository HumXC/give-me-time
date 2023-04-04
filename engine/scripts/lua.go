package scripts

import (
	"fmt"
	"time"

	"github.com/HumXC/give-me-time/engine/config"
	"github.com/Shopify/go-lua"
)

// 此文件用于定义在 lua 中使用的函数，函数名以 "luaFunc" 开头，后面的单词代表方法名，
// 但是在 lua 以小写开头的驼峰命名，例如 luaFuncPress 函数在 lua 中使用 "press" 调用。
// 该文件下的函数定义符合该目录下 exampel.lua 的描述，luaFunc 内需验证参数的正确性。
// 函数注释以一行是函数在 lua 中的使用方法
// 例如 press(element|x, duration|y, duration)
// 是指在第一个参数为 element 时（以 "press(E...)" 的形式调用）第二个参数为 duration 参数
// 当第一和第二参数为坐标 x 和 y 时，第三个参数为 duration

// sleep(duration) 暂停 duration 毫秒
func luaFuncSleep() lua.Function {
	return func(l *lua.State) (rt int) {
		duration, err := getDuration(l, 1)
		if err != nil {
			pushErr(l, err)
			return
		}
		time.Sleep(time.Duration(duration) * time.Millisecond)
		return
	}
}

// press(element|x, duration|y, duration)
// 按下屏幕上 element 或者坐标(x,y) 持续 duration 毫秒
func luaFuncPress(api Api, storage Storage) lua.Function {
	return func(l *lua.State) (rt int) {
		if ok, path := isElement(l, 1); ok {
			duration, err := getDuration(l, 2)
			if err != nil {
				pushErr(l, err)
				return
			}
			e := storage.Element(path)
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
	}
}

// swipe(element|x, |y).to(element|x, |y).action(duration)
// 在 duration 毫秒内从 swipe 传入的点滑动到 to 传入的点
func luaFuncSwipe(api Api, storage Storage) lua.Function {
	return func(l *lua.State) (rt int) {
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
				if ok, path := isElement(l, 1); ok {
					e := storage.Element(path)
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

		if ok, path := isElement(l, 1); ok {
			e := storage.Element(path)
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
	}
}

// find(element) (x, y, maxVal)
func luaFuncFind(api Api, storage Storage) lua.Function {
	return func(l *lua.State) int {
		if ok, path := isElement(l, 1); ok {
			e := storage.Element(path)
			if e.Img == "" {
				pushErr(l, fmt.Errorf("failed to find element[%s], it`s not a img element", path))
				return 0
			}
			p, v, err := api.FindE(e)
			if err != nil {
				pushErr(l, fmt.Errorf("failed to find element[%s]: %w", path, err))
				return 0
			}
			l.PushInteger(p.X)
			l.PushInteger(p.Y)
			l.PushNumber(float64(v))
			return 3
		}
		pushErr(l, fmt.Errorf("must be an element"))
		return 0
	}
}

// lock()
func luaFuncLock(api Api) lua.Function {
	return func(l *lua.State) (rt int) {
		err := api.Lock()
		if err != nil {
			pushErr(l, err)
		}
		return
	}
}

// unlock()
func luaFuncUnlock(api Api) lua.Function {
	return func(l *lua.State) (rt int) {
		err := api.Unlock()
		if err != nil {
			pushErr(l, err)
		}
		return
	}
}

// adb(cmd string) string
// 执行 adb 命令，adb 命令已经附加了 -s 参数
// 如果入参是 “shell ls”，实际执行的命令是“adb -s [...] shell ls”
func luaFuncAdb(api Api) lua.Function {
	return func(l *lua.State) (rt int) {
		rt = 1
		if l.TypeOf(1) != lua.TypeString {
			pushErr(l, fmt.Errorf("parameter error, [%v] not a string", l.ToValue(1)))
			return 0
		}
		cmd, _ := l.ToString(1)
		out, err := api.Adb(cmd)
		if err != nil {
			pushErr(l, err)
			return 0
		}
		l.PushString(string(out))
		return
	}
}

// ocr(x1, y1, x2, y2) string
// 返回范围内的文字识别结果
func luaFuncOcr(api Api) lua.Function {
	return func(l *lua.State) (rt int) {
		rt = 1
		x, y, err := getXY(l, 1, 2)
		if err != nil {
			pushErr(l, err)
			return 0
		}
		x2, y2, err := getXY(l, 3, 4)
		if err != nil {
			pushErr(l, err)
			return 0
		}
		text, err := api.Ocr(x, y, x2, y2)
		if err != nil {
			pushErr(l, err)
			return 0
		}
		l.PushString(text)
		return
	}
}

func pushElement(l *lua.State, name string, es []config.Element) {
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
		path := name + e.Name
		l.PushString(e.Name)
		l.NewTable()

		push("_path", path)
		push("_type", "element")

		pushElement(l, path, e.Element)

		l.SetTable(-3)
	}
}

func pushErr(l *lua.State, err error) {
	lua.Errorf(l, err.Error())
}

// 从第 index 个参数中获取 duration，其中 duration 是一个正整数，如果参数不符合则返回 error。
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
// 如果是，则返回 Element 的 "_path"
func isElement(L *lua.State, index int) (bool, string) {
	if L.TypeOf(index) != lua.TypeTable {
		return false, ""
	}
	L.Field(index, "_type")
	t, ok := L.ToString(-1)
	if !ok && t != "element" {
		return false, ""
	}
	L.Field(index, "_path")
	s, ok := L.ToString(-1)
	if !ok {
		return false, ""
	}
	return true, s
}
