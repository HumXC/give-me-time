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
	s.l.Register("sleep", luaFuncSleep())
	s.l.Register("press", luaFuncPress(api, s.storage))
	s.l.Register("swipe", luaFuncSwipe(api, s.storage))
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
