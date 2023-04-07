package scripts

import (
	"fmt"
	"path"

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
	l       *lua.State
	file    string
	storage Storage
	log     slog.Logger
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
func (s *script) setFunction(api Api, file string) {
	s.l.Register("sleep", luaFuncSleep(s.log))
	s.l.Register("press", luaFuncPress(s.log, api, s.storage))
	s.l.Register("swipe", luaFuncSwipe(s.log, api, s.storage))
	s.l.Register("find", luaFuncFind(s.log, api, s.storage))
	s.l.Register("lock", luaFuncLock(s.log, api))
	s.l.Register("unlock", luaFuncUnlock(s.log, api))
	s.l.Register("adb", luaFuncAdb(s.log, api))
	s.l.Register("ocr", luaFuncOcr(s.log, api))
	s.l.Register("read_json", luaFuncReadJson(s.log, path.Dir(file)))
}

// 设置在 lua 中的全局 E
func (s *script) setElement(es []config.Element) {
	s.l.NewTable()
	defer func() {
		s.l.SetGlobal("E")
	}()
	pushElement(s.l, "", es)
}

func LoadScript(log slog.Logger, file string, info *config.Info, element []config.Element, api Api) Script {
	s := &script{
		l:    lua.NewState(),
		file: file,
		storage: Storage{
			element: make(map[string]config.Element),
		},
		log: log,
	}
	config.FlatElement(s.storage.element, "", element)
	lua.OpenLibraries(s.l)
	s.setElement(element)
	s.setFunction(api, file)

	return s
}
