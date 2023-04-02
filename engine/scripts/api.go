package scripts

import (
	"errors"
	"fmt"
	"image"
	"os"

	"github.com/HumXC/give-me-time/cv"
	"github.com/HumXC/give-me-time/devices"
	"github.com/HumXC/give-me-time/engine/config"
	"gocv.io/x/gocv"
)

// “Element” 类型参数是指在 lua 中以 “click(E.main.start)” 的形式调用
type Api interface {
	// 按下一个元素或坐标，duration 单位是 ms。duration 为 0 时，将会自动把 duration 赋值为 100
	Press(x, y, duration int) error
	PressE(e config.Element, duration int) error
	// 滑动
	Swipe(x, y int) SwipeTo
	SwipeE(e config.Element) SwipeTo
	// 查找元素
	FindE(e config.Element) (image.Point, float32, error)
	// 锁定
	Lock() error
	Unlock() error
}

// Api 中的 Swipe 函数返回给 lua 一个 SwipeTo
type SwipeTo interface {
	To(x, y int) SwipeAction
	ToE(config.Element) SwipeAction
}
type SwipeAction interface {
	Action(duration int) error
}
type ApiImpl struct {
	Device  devices.Device
	Element map[string]config.Element
	// ElementMat 是配置了 Src 字段的 Element 对应的 gocv.Mat 实例
	ElementMat map[string]gocv.Mat
	Img        *gocv.Mat
}

func NewApi(device devices.Device, element []config.Element) (Api, error) {
	a := ApiImpl{
		Device:     device,
		Element:    make(map[string]config.Element),
		ElementMat: make(map[string]gocv.Mat),
	}
	config.FlatElement(a.Element, "", element)
	for k, e := range a.Element {
		_, err := os.Stat(e.Img)
		if err != nil {
			return nil, fmt.Errorf("can not get element[%s] src[%s] file stat: %w", e.Path, e.Img, err)
		}
		a.ElementMat[k] = gocv.IMRead(e.Img, gocv.IMReadUnchanged)
	}
	return &a, nil
}

func (a *ApiImpl) Press(x, y, duration int) error {
	return a.Device.Input.Press(x, y, duration)
}

func (a *ApiImpl) PressE(e config.Element, duration int) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to press element[%s]: %w", e.Path, err)
	}
	tmpl := a.ElementMat[e.Path]
	img, err := a.GetImg()
	if err != nil {
		return makeErr(err)
	}
	val, point, err := cv.Find(img, tmpl)
	if err != nil {
		return makeErr(err)
	}
	if val < 0.6 {
		// TODO 更详细的日志
		fmt.Printf("maxVal[%f] is too small\n", val)
	}

	err = a.Press(point.X+e.Offset.X, point.Y+e.Offset.Y, duration)
	if err != nil {
		return makeErr(err)
	}
	return nil
}

type SwipeHandler struct {
	api    *ApiImpl
	p1, p2 image.Point
	e1, e2 config.Element
}

func (h *SwipeHandler) To(x, y int) SwipeAction {
	h.p2.X = x
	h.p2.Y = y
	return h
}

func (h *SwipeHandler) ToE(e config.Element) SwipeAction {
	h.e2 = e
	return h
}

func (h *SwipeHandler) Action(duration int) error {
	var img *gocv.Mat
	makeErr := func(err error) error {
		// TODO 更详细的错误
		return fmt.Errorf("failed to swipe: %w", err)
	}

	find := func(el config.Element) (image.Point, error) {
		if img == nil {
			_img, err := h.api.GetImg()
			if err != nil {
				return image.ZP, makeErr(err)
			}
			img = &_img
		}
		tmpl := h.api.ElementMat[el.Path]
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
	err := h.api.Device.Input.Swipe(h.p1.X, h.p1.Y, h.p2.X, h.p2.Y, duration)
	if err != nil {
		return makeErr(err)
	}
	return nil
}

func (a *ApiImpl) Swipe(x, y int) SwipeTo {
	return &SwipeHandler{
		api: a,
		p1:  image.Point{X: x, Y: y},
	}
}

func (a *ApiImpl) SwipeE(e config.Element) SwipeTo {
	return &SwipeHandler{
		api: a,
		e1:  e,
	}
}

func (a *ApiImpl) FindE(e config.Element) (image.Point, float32, error) {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to find element[%s]: %w", e.Path, err)
	}
	tmpl := a.ElementMat[e.Path]
	img, err := a.GetImg()
	if err != nil {
		return image.ZP, 0, makeErr(err)
	}
	val, point, err := cv.Find(img, tmpl)
	if err != nil {
		return image.ZP, 0, makeErr(err)
	}
	return image.Pt(point.X+e.Offset.X, point.Y+e.Offset.Y), val, nil
}
func (a *ApiImpl) GetImg() (gocv.Mat, error) {
	if a.Img != nil {
		return *a.Img, nil
	}
	imgB, err := a.Device.ADB("shell screencap -p")
	if err != nil {
		return gocv.NewMat(), err
	}
	img, err := gocv.IMDecode(imgB, gocv.IMReadUnchanged)
	if err != nil {
		return gocv.NewMat(), err
	}
	return img, nil
}
func (a *ApiImpl) Lock() error {
	if a.Img != nil {
		return errors.New("can not be locked repeatedly")
	}
	img, err := a.GetImg()
	if err != nil {
		return fmt.Errorf("failed to lock: %w", err)
	}
	a.Img = &img
	return nil
}
func (a *ApiImpl) Unlock() error {
	if a.Img == nil {
		return errors.New("can not be unlocked repeatedly")
	}
	a.Img = nil
	return nil
}
