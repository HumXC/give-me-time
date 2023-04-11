package scripts

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/HumXC/give-me-time/engine/config"
	"github.com/Shopify/go-lua"
	"golang.org/x/exp/slog"
)

// 此文件用于定义在 lua 中使用的函数，连接 api.go 中定义的函数。

// 关于函数名：
// 函数名以 "luaFunc" 开头，后面的单词代表方法名，
// 但是在 lua 以下划线命名法命名，例如 luaFuncPress 函数在 lua 中使用 "press" 调用。
// 该文件下的函数定义符合该目录下 exampel.lua 的描述，luaFunc 内需验证参数的正确性。
// 函数注释以一行是函数在 lua 中的使用方法
// 例如 press(element|x, duration|y, duration)
// 是指在第一个参数为 element 时（以 "press(E...)" 的形式调用）第二个参数为 duration 参数
// 当第一和第二参数为坐标 x 和 y 时，第三个参数为 duration

// sleep(duration) 暂停 duration 毫秒
func luaFuncSleep(log slog.Logger) lua.Function {
	return func(l *lua.State) (rt int) {
		duration, err := getDuration(l, 1)
		if err != nil {
			pushErr(log, l, err)
			return
		}
		time.Sleep(time.Duration(duration) * time.Millisecond)
		log.Info(fmt.Sprintf("sleep %d millisecond", duration))
		return
	}
}

// press(element|x, duration|y, duration) ok
// 按下屏幕上 element 或者坐标(x,y) 持续 duration 毫秒
func luaFuncPress(log slog.Logger, api Api, storage Storage) lua.Function {
	return func(l *lua.State) int {
		if ok, path := isElement(l, 1); ok {
			duration, err := getDuration(l, 2)
			if err != nil {
				pushErr(log, l, err)
				return 0
			}
			e := storage.Element(path)
			ok, err = api.PressE(e, duration)
			if err != nil {
				pushErr(log, l, err)
				return 0
			}
			if !ok {
				log.Info(fmt.Sprintf("failed to press element [%s] use %d millisecond, [maxVal] is smaller than [threshold:%f]",
					e.Name, duration, e.Threshold))
				l.PushBoolean(ok)
				return 1
			}
			log.Info(fmt.Sprintf("press element [%s] %d millisecond", e.Name, duration))
			l.PushBoolean(ok)
			return 1
		}
		duration, err := getDuration(l, 3)
		if err != nil {
			pushErr(log, l, err)
			return 0
		}
		x, y, err := getXY(l, 1, 2)
		if err != nil {
			pushErr(log, l, err)
			return 0
		}
		err = api.Press(x, y, duration)
		if err != nil {
			pushErr(log, l, err)
			return 0
		}
		log.Info(fmt.Sprintf("press (%d, %d) %d millisecond", x, y, duration))
		l.PushBoolean(true)
		return 1
	}
}

// swipe(element|x, |y).to(element|x, |y).action(duration)
// 在 duration 毫秒内从 swipe 传入的点滑动到 to 传入的点
func luaFuncSwipe(log slog.Logger, api Api, storage Storage) lua.Function {
	return func(l *lua.State) int {
		setSwipeAction := func(l *lua.State, st SwipeAction) int {
			l.NewTable()
			l.PushString("action")
			l.PushGoFunction(func(state *lua.State) int {
				duration, err := getDuration(l, 1)
				if err != nil {
					pushErr(log, l, err)
					return 0
				}
				p1, p2, ok, err := st.Action(duration)
				if err != nil {
					pushErr(log, l, err)
					return 0
				}
				if !ok {
					log.Info(fmt.Sprintf(
						"failed to swipe (%d, %d) to (%d, %d) use %d millisecond, there are some points that do not reach the threshold",
						p1.X, p1.Y, p2.X, p2.Y, duration))
					l.PushBoolean(false)
					return 1
				}
				log.Info(fmt.Sprintf(
					"swipe (%d, %d) to (%d, %d) use %d millisecond",
					p1.X, p1.Y, p2.X, p2.Y, duration))
				l.PushBoolean(true)
				return 1
			})
			l.SetTable(-3)
			return 1
		}
		setSwipeTo := func(l *lua.State, st SwipeTo) int {
			l.NewTable()
			l.PushString("to")
			l.PushGoFunction(func(state *lua.State) int {
				if ok, path := isElement(l, 1); ok {
					e := storage.Element(path)
					sac := st.ToE(e)
					setSwipeAction(l, sac)
					return 1
				}
				x, y, err := getXY(l, 1, 2)
				if err != nil {
					pushErr(log, l, err)
					return 0
				}
				sac := st.To(x, y)
				setSwipeAction(l, sac)
				return 1
			})
			l.SetTable(-3)
			return 1
		}

		if ok, path := isElement(l, 1); ok {
			e := storage.Element(path)
			st := api.SwipeE(e)
			setSwipeTo(l, st)
			return 1
		}
		x, y, err := getXY(l, 1, 2)
		if err != nil {
			pushErr(log, l, err)
			return 0
		}
		st := api.Swipe(x, y)
		setSwipeTo(l, st)
		return 1
	}
}

// find(element) (x, y, maxVal)
func luaFuncFind(log slog.Logger, api Api, storage Storage) lua.Function {
	return func(l *lua.State) int {
		if ok, path := isElement(l, 1); ok {
			e := storage.Element(path)
			if e.Img == "" {
				pushErr(log, l, fmt.Errorf("element [%s] does not exist", path))
				return 0
			}
			p, v, err := api.FindE(e)
			if err != nil {
				pushErr(log, l, fmt.Errorf("element [%s] not found: %w", path, err))
				return 0
			}
			log.Info(fmt.Sprintf(
				"find element [%s] on (%d, %d), val: %f", e.Name, p.X, p.Y, v))
			l.PushInteger(p.X)
			l.PushInteger(p.Y)
			l.PushNumber(float64(v))
			return 3
		}
		pushErr(log, l, fmt.Errorf("must be an element"))
		return 0
	}
}

// lock()
func luaFuncLock(log slog.Logger, api Api) lua.Function {
	return func(l *lua.State) (rt int) {
		err := api.Lock()
		if err != nil {
			pushErr(log, l, err)
		}
		log.Info("lock")
		return
	}
}

// unlock()
func luaFuncUnlock(log slog.Logger, api Api) lua.Function {
	return func(l *lua.State) (rt int) {
		err := api.Unlock()
		if err != nil {
			pushErr(log, l, err)
		}
		log.Info("unlock")
		return
	}
}

// adb(cmd string) string
// 执行 adb 命令，adb 命令已经附加了 -s 参数
// 如果入参是 “shell ls”，实际执行的命令是“adb -s [...] shell ls”
func luaFuncAdb(log slog.Logger, api Api) lua.Function {
	return func(l *lua.State) (rt int) {
		rt = 1
		if l.TypeOf(1) != lua.TypeString {
			pushErr(log, l, fmt.Errorf("the parameter [%v] is not a string", l.ToValue(1)))
			return 0
		}
		cmd, _ := l.ToString(1)
		out, err := api.Adb(cmd)
		if err != nil {
			pushErr(log, l, fmt.Errorf("adb error: %w", err))
			return 0
		}
		log.Info("run adb: [" + cmd + "]")
		l.PushString(string(out))
		return
	}
}

// ocr(x1, y1, x2, y2) string
// 返回范围内的文字识别结果
func luaFuncOcr(log slog.Logger, api Api) lua.Function {
	return func(l *lua.State) (rt int) {
		rt = 1
		x, y, err := getXY(l, 1, 2)
		if err != nil {
			pushErr(log, l, err)
			return 0
		}
		x2, y2, err := getXY(l, 3, 4)
		if err != nil {
			pushErr(log, l, err)
			return 0
		}
		text, err := api.Ocr(x, y, x2, y2)
		if err != nil {
			pushErr(log, l, err)
			return 0
		}
		log.Info(fmt.Sprintf("ocr (%d, %d)-(%d, %d): %s", x, y, x2, y2, text))
		l.PushString(text)
		return
	}
}

// read_json(file) table
// 读取一个 json 文件
func luaFuncReadJson(log slog.Logger, dir string) lua.Function {
	return func(l *lua.State) int {
		fileName, ok := l.ToString(1)
		if !ok {
			pushErr(log, l, errors.New("file must be a string"))
			return 0
		}
		file := PatchAbsPath(fileName, dir)
		b, err := os.ReadFile(file)
		if err != nil {
			pushErr(log, l, fmt.Errorf("read file error: %w", err))
			return 0
		}
		m := make(map[string]any)
		err = json.Unmarshal(b, &m)
		if err != nil {
			pushErr(log, l, fmt.Errorf("json unmarshal error: %w", err))
			return 0
		}
		PushMap(l, m)
		return 1
	}
}

// PatchAbsPath
func PatchAbsPath(p, dir string) string {
	if !path.IsAbs(p) {
		return path.Join(dir, p)
	}
	return p
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

func pushErr(log slog.Logger, l *lua.State, err error) {
	msg := "lua bound function call error: " + err.Error()
	log.Error(msg)
	lua.Errorf(l, msg)
}

// 从第 index 个参数中获取 duration，其中 duration 是一个正整数，如果参数不符合则返回 error。
func getDuration(l *lua.State, index int) (int, error) {
	v := l.ToValue(index)
	if v == nil {
		return 100, nil
	}
	d, ok := l.ToInteger(index)
	if !ok {
		return 0, fmt.Errorf("the duration [%v] is not an integer", v)
	}
	if d < 0 {
		return 0, fmt.Errorf("the duration [%d] is not a positive integer", d)
	}
	if d == 0 {
		d = 100
	}
	return d, nil
}

// 获取两个整数作为“坐标”使用
func getXY(l *lua.State, indexX, indexY int) (int, int, error) {
	x, ok := l.ToInteger(indexX)
	if !ok {
		err := fmt.Errorf("the x [%v] is not an integer", l.ToValue(1))
		return 0, 0, err
	}
	y, ok := l.ToInteger(indexY)
	if !ok {
		err := fmt.Errorf("the y [%v] is not an integer", l.ToValue(2))
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
