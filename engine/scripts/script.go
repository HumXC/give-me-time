package scripts

import (
	"fmt"

	"github.com/HumXC/give-me-time/engine/config"
	"github.com/Shopify/go-lua"
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
	s.l.Register("sleep", luaFuncSleep())
	s.l.Register("press", luaFuncPress(api, s.storage))
	s.l.Register("swipe", luaFuncSwipe(api, s.storage))
	s.l.Register("find", luaFuncFind(api, s.storage))
	s.l.Register("lock", luaFuncLock(api))
	s.l.Register("unlock", luaFuncUnlock(api))
	s.l.Register("adb", luaFuncAdb(api))
}

// 设置在 lua 中的全局 E
func (s *script) setElement(es []config.Element) {
	s.l.NewTable()
	defer func() {
		s.l.SetGlobal("E")
	}()
	pushElement(s.l, "", es)
}

func LoadScript(file string, info *config.Info, element []config.Element, api Api) Script {
	s := &script{
		l:    lua.NewState(),
		file: file,
		storage: Storage{
			element: make(map[string]config.Element),
		},
	}
	config.FlatElement(s.storage.element, "", element)
	lua.OpenLibraries(s.l)
	s.setElement(element)
	s.setFunction(api)

	return s
}