package engine

import (
	"fmt"
	"image"
	"os"

	"github.com/HumXC/give-me-time/cv"
	"github.com/HumXC/give-me-time/devices"
	"gocv.io/x/gocv"
)

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
	Device  devices.Device
	Element map[string]Element
	// ElementMat 是配置了 Src 字段的 Element 对应的 gocv.Mat 实例
	ElementMat map[string]gocv.Mat
}

func NewApi(device devices.Device, element []Element) (Api, error) {
	a := ApiImpl{
		Device:     device,
		Element:    make(map[string]Element),
		ElementMat: make(map[string]gocv.Mat),
	}
	FlatElement(a.Element, "", element)
	for k, e := range a.Element {
		if e.Src == "" {
			continue
		}
		_, err := os.Stat(e.Src)
		if err != nil {
			return nil, fmt.Errorf("can not get element[%s] src[%s] file stat: %w", e.Path, e.Src, err)
		}
		a.ElementMat[k] = gocv.IMRead(e.Src, gocv.IMReadUnchanged)
	}
	return &a, nil
}
func (a *ApiImpl) Press(x, y, duration int) error {
	return a.Device.Input.Press(x, y, duration)
}
func (a *ApiImpl) PressE(e Element, duration int) error {
	tmpl := a.ElementMat[e.Path]
	imgB, err := a.Device.Screenshot()
	if err != nil {
		return fmt.Errorf("failed to press element[%s]: %w", e.Path, err)
	}
	img, err := gocv.IMDecode(imgB, gocv.IMReadUnchanged)
	if err != nil {
		return fmt.Errorf("failed to press element[%s]: %w", e.Path, err)
	}
	val, point, err := cv.Find(img, tmpl)
	if err != nil {
		return fmt.Errorf("failed to press element[%s]: %w", e.Path, err)
	}
	if val < 0.6 {
		// TODO 更详细的日志
		fmt.Printf("maxVal[%f] is too small\n", val)
	}

	var p = image.Pt(point.X+e.Offset.X, point.Y+e.Offset.Y)
	err = a.Press(p.X, p.Y, duration)
	if err != nil {
		return fmt.Errorf("failed to press element[%s]: %w", e.Path, err)
	}
	return nil
}

type SwipeHandler struct {
	device     devices.Device
	elementMat map[string]gocv.Mat
	p1, p2     image.Point
	e1, e2     Element
}

func (h *SwipeHandler) To(x, y int) SwipeAction {
	h.p2.X = x
	h.p2.Y = y
	return h
}
func (h *SwipeHandler) ToE(e Element) SwipeAction {
	h.e2 = e
	return h
}
func (h *SwipeHandler) Action(duration int) error {
	var img *gocv.Mat
	makeErr := func(err error) error {
		// TODO 更详细的错误
		return fmt.Errorf("failed to swipe: %w", err)
	}

	find := func(el Element) (image.Point, error) {
		if img == nil {
			b, err := h.device.Screenshot()
			if err != nil {
				return image.ZP, err
			}
			_img, err := gocv.IMDecode(b, gocv.IMReadUnchanged)
			if err != nil {
				return image.ZP, err
			}
			img = &_img
		}
		tmpl := h.elementMat[el.Path]
		val, point, err := cv.Find(*img, tmpl)
		if val < 0.6 {
			// TODO 更详细的日志
			fmt.Printf("maxVal[%f] is too small\n", val)
		}
		if err != nil {
			return point, err
		}
		point.X += el.Offset.X
		point.Y += el.Offset.Y
		return point, nil
	}
	if h.e1.Name != "" {
		p, err := find(h.e1)
		if err != nil {
			return makeErr(err)
		}
		h.p1 = p
	}
	if h.e2.Name != "" {
		p, err := find(h.e2)
		if err != nil {
			return makeErr(err)
		}
		h.p1 = p
	}
	err := h.device.Input.Swipe(h.p1.X, h.p1.Y, h.p2.X, h.p2.Y, duration)
	if err != nil {
		return makeErr(err)
	}
	return nil
}
func (a *ApiImpl) Swipe(x, y int) SwipeTo {
	return &SwipeHandler{
		elementMat: a.ElementMat,
		device:     a.Device,
		p1:         image.Point{X: x, Y: y},
	}
}
func (a *ApiImpl) SwipeE(e Element) SwipeTo {
	return &SwipeHandler{
		elementMat: a.ElementMat,
		device:     a.Device,
		e1:         e,
	}
}
