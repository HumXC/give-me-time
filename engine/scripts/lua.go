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

// sleep(duration) 暂停 duration 毫秒
func luaFuncSleep(log slog.Logger) lua.Function {
	return func(l *lua.State) (rt int) {
		duration, ok := l.ToInteger(1)
		if !ok {
			PushErr(log, l, NewArgsErr("number", l.ToValue(1)))
			return
		}
		if duration < 0 {
			PushErr(log, l, NewArgsErr("number bigger than 0", duration))
			return
		}
		time.Sleep(time.Duration(duration) * time.Millisecond)
		log.Info(fmt.Sprintf("sleep %d millisecond", duration))
		return
	}
}

// read_json(file) table
// 读取一个 json 文件
func luaFuncReadJson(log slog.Logger, dir string) lua.Function {
	return func(l *lua.State) int {
		fileName, ok := l.ToString(1)
		if !ok {
			PushErr(log, l, errors.New("file must be a string"))
			return 0
		}
		file := PatchAbsPath(fileName, dir)
		b, err := os.ReadFile(file)
		if err != nil {
			PushErr(log, l, fmt.Errorf("read file error: %w", err))
			return 0
		}
		m := make(map[string]any)
		err = json.Unmarshal(b, &m)
		if err != nil {
			PushErr(log, l, fmt.Errorf("json unmarshal error: %w", err))
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
func PushElement(l *lua.State, name string, es []config.Element) {
	if len(es) == 0 {
		return
	}
	if name != "" {
		name += "."
	}
	pushString := func(k, v string) {
		l.PushString(k)
		l.PushString(v)
		l.SetTable(-3)
	}
	for _, e := range es {
		path := name + e.Name
		l.PushString(e.Name)
		l.NewTable()

		pushString("_path", path)
		pushString("_type", "element")
		pushString("discription", e.Discription)
		switch e.Type {
		case config.ElTypeImg:
			pushString("img", e.Img)
			l.PushString("threshold")
			l.PushNumber(float64(e.Threshold))
			l.SetTable(-3)
			l.PushString("offset")
			PushMap(l, map[string]any{"x": e.Offset.X, "y": e.Offset.Y})
			l.SetTable(-3)
		case config.ElTypeArea:
			l.PushString("x1")
			l.PushInteger(e.Area.X1)
			l.SetTable(-3)
			l.PushString("y1")
			l.PushInteger(e.Area.Y1)
			l.SetTable(-3)
			l.PushString("x2")
			l.PushInteger(e.Area.X2)
			l.SetTable(-3)
			l.PushString("y2")
			l.PushInteger(e.Area.Y2)
			l.SetTable(-3)
		case config.ElTypePoint:
			l.PushString("x")
			l.PushInteger(e.Point.X)
			l.SetTable(-3)
			l.PushString("y")
			l.PushInteger(e.Point.Y)
			l.SetTable(-3)
		}

		PushElement(l, path, e.Element)

		l.SetTable(-3)
	}
}

func PushErr(log slog.Logger, l *lua.State, err error) {
	msg := "lua bound function call error: " + err.Error()
	log.Error(msg)
	lua.Errorf(l, msg)
}
