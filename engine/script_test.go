package engine_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/HumXC/give-me-time/engine"
	"github.com/Shopify/go-lua"
)

func TestLoadScript(t *testing.T) {
	opt, err := engine.LoadOption("test.json")
	if err != nil {
		t.Error(err)
		return
	}
	s := engine.LoadScript("test.lua", opt)
	err = s.Run()
	if err != nil {
		t.Error(err)
	}
}

// 以下内容是试验代码，不要作为测试运行
type TestSwiper struct {
	data string
}

func (t TestSwiper) To(L *lua.State) (rt int) {
	rt = 1
	defer func() {
		L.NewTable()
		L.PushString("to")
		L.PushGoFunction(t.To)
		L.SetTable(-3)
		L.PushString("action")
		L.PushGoFunction(t.Action)
		L.SetTable(-3)
	}()
	if L.TypeOf(1) == lua.TypeTable {
		L.Field(1, "_name")
		s, ok := L.ToString(-1)
		if !ok {
			fmt.Println("error")
			return
		}
		t.data += " 到 " + s
		return
	}
	x, ok := L.ToInteger(1)
	y, ok2 := L.ToInteger(2)
	if !ok || !ok2 {
		fmt.Println("error")
		return
	}
	t.data += fmt.Sprintf("从 [%d, %d]", x, y)
	return
}
func (t TestSwiper) Action(L *lua.State) int {
	fmt.Println(t.data)
	return 0
}

type TestApi struct{}

func (t TestApi) Press(L *lua.State) int {
	fmt.Println("press")
	return 0
}
func (t TestApi) Swipe(L *lua.State) (rt int) {
	rt = 1
	fmt.Println("swipe")
	var swiper TestSwiper
	defer func() {
		L.NewTable()
		L.PushString("to")
		L.PushGoFunction(swiper.To)
		L.SetTable(-3)
		L.PushString("action")
		L.PushGoFunction(swiper.Action)
		L.SetTable(-3)
	}()
	if L.TypeOf(1) == lua.TypeTable {
		L.Field(1, "_name")
		s, ok := L.ToString(-1)
		if !ok {
			fmt.Println("error")
			return
		}
		swiper = TestSwiper{
			data: "从 " + s,
		}
		return
	}
	x, ok := L.ToInteger(1)
	y, ok2 := L.ToInteger(2)
	if !ok || !ok2 {
		fmt.Println("error")
		return
	}
	swiper = TestSwiper{
		data: fmt.Sprintf("从 [%d, %d]", x, y),
	}
	return
}
func (t TestApi) Ocr(L *lua.State) int {
	fmt.Println("ocr")
	return 0
}
func (t TestApi) Find(L *lua.State) int {
	fmt.Println("find")
	return 0
}

func TestXsxx(t *testing.T) {
	var api engine.Api
	api = TestApi{}
	L := lua.NewState()
	lua.OpenLibraries(L)

	L.Register("press", api.Press)
	L.Register("swipe", api.Swipe)
	err := lua.DoFile(L, "test.lua")
	if err != nil {
		log.Print(err)
	}
}
