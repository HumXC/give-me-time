package scripts

import (
	"fmt"
	"image"

	"github.com/HumXC/adb-helper"
	"github.com/Shopify/go-lua"
	"golang.org/x/exp/slog"
)

type ApiAdb interface {
	LuaApi
	InputHandler
	// 执行 adb 命令
	Cmd(string) ([]byte, error)
}

type apiAdbImpl struct {
	input adb.Input
	cmd   adb.ADBRunner
}

func (a *apiAdbImpl) Cmd(cmd string) ([]byte, error) {
	out, err := a.cmd(cmd)
	if err != nil {
		return nil, fmt.Errorf("adb error: %w", err)
	}
	return out, nil
}

func (a *apiAdbImpl) Press(x, y, duration int) error {
	return a.input.Press(x, y, duration)
}

type SwipeHandler struct {
	swipe  func(x1, y1, x2, y2, duration int) error
	p1, p2 image.Point
}

func (h *SwipeHandler) To(x, y int) InputHandlerSwipeAction {
	h.p2.X = x
	h.p2.Y = y
	return h
}

func (h *SwipeHandler) Action(duration int) (image.Point, image.Point, bool, error) {

	err := h.swipe(h.p1.X, h.p1.Y, h.p2.X, h.p2.Y, duration)
	if err != nil {
		return image.ZP, image.ZP, false, err
	}
	return h.p1, h.p2, true, nil
}

func (a *apiAdbImpl) Swipe(x, y int) InputHandlerSwipeTo {
	return &SwipeHandler{
		swipe: a.input.Swipe,
		p1:    image.Point{X: x, Y: y},
	}
}

func (a *apiAdbImpl) ToLuaFunc(log slog.Logger) map[string]lua.Function {
	m := make(map[string]lua.Function)
	// adb(cmd string) string
	// 执行 adb 命令，adb 命令已经附加了 -s 参数
	// 如果入参是 “shell ls”，实际执行的命令是“adb -s [...] shell ls”
	m["cmd"] = func(l *lua.State) int {
		args := NewArgsPicker(l)
		s, ok := args.StringWithNotEmpty(1)
		if !ok {
			PushErr(log, l, NewArgsErr("string", l.ToValue(1)))
			return 0
		}
		result, err := a.Cmd(s)
		if err != nil {
			PushErr(log, l, err)
			return 0
		}
		l.PushString(string(result))
		log.Info(fmt.Sprintf("adb cmd [%s]", s))
		return 1
	}
	// press(x, y, duration int)
	m["press"] = func(l *lua.State) int {
		args := NewArgsPicker(l)
		x, ok := args.Int(1)
		if !ok {
			PushErr(log, l, NewArgsErr("number", l.ToValue(1)))
			return 0
		}
		y, ok := args.Int(2)
		if !ok {
			PushErr(log, l, NewArgsErr("number", l.ToValue(2)))
			return 0
		}
		duration, ok := args.Int(3)
		if !ok {
			PushErr(log, l, NewArgsErr("number", l.ToValue(3)))
			return 0
		}
		err := a.Press(x, y, duration)
		if err != nil {
			PushErr(log, l, fmt.Errorf("failed to press (%d, %d) %d ms: %w", x, y, duration, err))
			return 0
		}
		log.Info(fmt.Sprintf("press (%d, %d) %d ms", x, y, duration))
		return 0
	}
	// swipe(x, y int).to(x, y int).action(duration int)
	m["swipe"] = func(l *lua.State) int {
		args := NewArgsPicker(l)
		setSwipeAction := func(l *lua.State, st InputHandlerSwipeAction) int {
			l.NewTable()
			l.PushString("action")
			l.PushGoFunction(func(state *lua.State) int {
				duration, ok := args.IntWithBigger(1, 0)
				if !ok {
					PushErr(log, l, NewArgsErr("number", l.ToValue(1)))
					return 0
				}
				p1, p2, ok, err := st.Action(duration)
				if err != nil {
					PushErr(log, l, err)
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
		setSwipeTo := func(_x, _y int, l *lua.State, st InputHandlerSwipeTo) int {
			l.NewTable()
			l.PushString("to")
			l.PushGoFunction(func(state *lua.State) int {
				x, ok := args.Int(1)
				if !ok {
					PushErr(log, l, NewArgsErr("number", l.ToValue(1)))
					return 0
				}
				y, ok := args.Int(2)
				if !ok {
					PushErr(log, l, NewArgsErr("number", l.ToValue(2)))
					return 0
				}
				sac := st.To(x, y)
				setSwipeAction(l, sac)
				log.Info(fmt.Sprintf("set swipe (%d, %d) to (%d, %d)", _x, _y, x, y))
				return 1
			})
			l.SetTable(-3)
			return 1
		}
		x, ok := args.Int(1)
		if !ok {
			PushErr(log, l, NewArgsErr("number", l.ToValue(1)))
			return 0
		}
		y, ok := args.Int(2)
		if !ok {
			PushErr(log, l, NewArgsErr("number", l.ToValue(2)))
			return 0
		}
		st := a.Swipe(x, y)
		setSwipeTo(x, y, l, st)
		log.Info(fmt.Sprintf("set swipe (%d, %d)", x, y))
		return 1
	}
	return m
}
func NewApiAdb(device adb.Device) ApiAdb {
	return &apiAdbImpl{
		input: device.Input,
		cmd:   device.Cmd,
	}
}
