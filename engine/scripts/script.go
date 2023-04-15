package scripts

import (
	"fmt"

	"github.com/HumXC/adb-helper"
	"github.com/HumXC/give-me-time/engine/config"
	"github.com/Shopify/go-lua"
	"golang.org/x/exp/slog"
)

type Script interface {
	// 开始运行脚本
	Run() error
	// 返回文件名
	File() string
}
type Storage struct {
	element map[string]config.Element
}

func (s *Storage) Element(key string) config.Element {
	return s.element[key]
}

type script struct {
	l    *lua.State
	file string
	log  slog.Logger
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
func (s *script) setFunction(api ApiImg, file string) {

}
func (s *script) setApi(name string, api LuaApi, log slog.Logger) {
	fns := api.ToLuaFunc(log)
	s.l.NewTable()
	for k, f := range fns {
		s.l.PushString(k)
		s.l.PushGoFunction(f)
		s.l.SetTable(-3)
	}
	s.l.SetGlobal(name)
}

// 设置在 lua 中的全局 E
func (s *script) setElement(es []config.Element) {
	s.l.NewTable()
	defer func() {
		s.l.SetGlobal("E")
	}()
	PushElement(s.l, "", es)
}

func LoadScript(device adb.Device, log slog.Logger, file string, info *config.Info, element []config.Element) (Script, error) {
	s := &script{
		l:    lua.NewState(),
		file: file,
		log:  log,
	}
	lua.OpenLibraries(s.l)
	s.setElement(element)
	elImg, elArea, _, err := config.ParseElement(element)
	apiImg, err := NewApiImg(device.Cmd, elImg, elArea)
	if err != nil {
		return nil, err
	}
	apiAdb := NewApiAdb(device)
	s.setApi("Img", apiImg, log)
	s.setApi("Adb", apiAdb, log)
	s.l.Register("sleep", luaFuncSleep(log))
	return s, nil
}
