package engine

import "github.com/HumXC/give-me-time/devices"

// “Element” 类型参数是指在 lua 中以 “click(E.main.start)” 的形式调用
type Api interface {
	// 按下一个元素或坐标，duration 单位是 ms。duration 为 0 时，将会自动把 duration 赋值为 100
	Press(x, y, duration int) error
	PressE(e Element, duration int) error
	// 滑动
	Swipe(x, y int) SwipeTo
	SwipeE(Element) SwipeTo
}

// Api 中的 Swipe 函数返回给 lua 一个 SwipeTo
type SwipeTo interface {
	To(x, y int) SwipeAction
	ToE(Element) SwipeAction
}
type SwipeAction interface {
	Action(duration int) error
}
type ApiImpl struct {
	Input devices.Input
}

func (a *ApiImpl) Press(x, y, duration int) error {
	return a.Input.Press(x, y, duration)
}
func (a *ApiImpl) PressE(e Element, duration int) error {
	// TODO 通过 CV 识别 e 的位置然后点击
	return nil
}

type SwipeHandler struct {
	input          devices.Input
	x1, y1, x2, y2 int
	e1, e2         Element
}

func (h *SwipeHandler) To(x, y int) SwipeAction {
	h.x2 = x
	h.y2 = y
	return h
}
func (h *SwipeHandler) ToE(e Element) SwipeAction {
	h.e2 = e
	return h
}
func (h *SwipeHandler) Action(duration int) error {
	if h.e1.Name == "" && h.e2.Name == "" {
		return h.input.Swipe(h.x1, h.y1, h.x2, h.y2, duration)
	}
	// TODO 通过 CV 识别
	return nil
}
func (a *ApiImpl) Swipe(x, y int) SwipeTo {
	return &SwipeHandler{input: a.Input, x1: x, y1: y}
}
func (a *ApiImpl) SwipeE(e Element) SwipeTo {
	return &SwipeHandler{input: a.Input, e1: e}
}
