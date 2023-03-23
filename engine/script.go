package engine

import "github.com/Shopify/go-lua"

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

type Api interface {
	Press(*lua.State) int
	Swipe(*lua.State) int
	Ocr(*lua.State) int
	Find(*lua.State) int
}

type Swiper interface {
	To(*lua.State) int
	Action(*lua.State) int
}

func LoadScript(file string, option *Option) Script {
	L := lua.NewState()
	lua.OpenLibraries(L)
	setElement(L, option.Element)
	return script{
		l:    L,
		file: file,
	}
}

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
	}
	for _, e := range es {
		_name := name + e.Name
		l.PushString(e.Name)

		l.NewTable()
		push("_name", _name)
		l.SetTable(-3)

		pushElement(l, _name, e.Element)
		l.SetTable(-3)
	}
}
