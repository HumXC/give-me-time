package scripts_test

import (
	"fmt"
	"image"
	"testing"

	"github.com/HumXC/give-me-time/engine/config"
	"github.com/HumXC/give-me-time/engine/scripts"
	"github.com/Shopify/go-lua"
)

type Api struct{}

type SwipeHandler struct{ data string }

func (h *SwipeHandler) To(x, y int) scripts.SwipeAction {
	h.data += fmt.Sprintf(" to (%d, %d)", x, y)
	return h
}
func (h *SwipeHandler) ToE(e config.Element) scripts.SwipeAction {
	h.data += fmt.Sprintf(" to (%s)", e.Discription)
	return h
}
func (h *SwipeHandler) Action(duration int) error {
	h.data += fmt.Sprintf(" use %d ms.", duration)
	return nil
}

func (a *Api) Press(x, y, d int) error {
	return nil
}
func (a *Api) PressE(e config.Element, d int) error {
	return nil
}
func (a *Api) Swipe(x, y int) scripts.SwipeTo {
	return &SwipeHandler{
		data: fmt.Sprintf("from (%d, %d)", x, y),
	}
}
func (a *Api) SwipeE(e config.Element) scripts.SwipeTo {
	return &SwipeHandler{
		data: fmt.Sprintf("from (%s)", e.Name),
	}
}
func (a *Api) FindE(e config.Element) (image.Point, float32, error) {
	return image.ZP, 0, nil
}
func (a *Api) Lock() error {
	return nil
}
func (a *Api) Unlock() error {
	return nil
}
func TestLoadScript(t *testing.T) {
	opt, err := config.LoadInfo("test.json")
	if err != nil {
		t.Error(err)
		return
	}
	elm, err := config.LoadElement("test.json")
	if err != nil {
		t.Error(err)
		return
	}
	var api scripts.Api
	api = &Api{}
	s := scripts.LoadScript("test.lua", opt, elm, api)
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

func (t TestApi) Click(L *lua.State) int {
	fmt.Println("click")
	return 0
}

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
