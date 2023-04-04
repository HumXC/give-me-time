package scripts

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"sync"

	"github.com/HumXC/give-me-time/cv"
	"github.com/HumXC/give-me-time/devices"
	"github.com/HumXC/give-me-time/engine/config"
	"github.com/otiai10/gosseract/v2"
	"gocv.io/x/gocv"
)

// “Element” 类型参数是指在 lua 中以 “click(E.main.start)” 的形式调用
type Api interface {
	// 按下一个元素或坐标，duration 单位是 ms。duration 为 0 时，将会自动把 duration 赋值为 100
	Press(x, y, duration int) error
	PressE(e config.Element, duration int) (bool, error)
	// 滑动
	Swipe(x, y int) SwipeTo
	SwipeE(e config.Element) SwipeTo
	// 查找元素
	FindE(e config.Element) (image.Point, float32, error)
	// 锁定与解锁当前 Find 函数的对象
	Lock() error
	Unlock() error
	// 执行 adb 命令
	Adb(string) ([]byte, error)
	Ocr(x1, y1, x2, y2 int) (string, error)
}

// Api 中的 Swipe 函数返回给 lua 一个 SwipeTo
type SwipeTo interface {
	To(x, y int) SwipeAction
	ToE(config.Element) SwipeAction
}
type SwipeAction interface {
	// 第一个返回值是开始滑动的点，第二个返回值是滑动结束的点，第三个返回值表示是否成功
	Action(duration int) (image.Point, image.Point, bool, error)
}
type ApiImpl struct {
	Device  devices.Device
	Element map[string]config.Element
	// ElementMat 是配置了 Src 字段的 Element 对应的 gocv.Mat 实例
	ElementMat map[string]gocv.Mat
	Img        *gocv.Mat
	Sseract    *gosseract.Client
	sseractMu  *sync.Mutex
}

func NewApi(device devices.Device, element []config.Element) (Api, error) {
	a := ApiImpl{
		Device:     device,
		Element:    make(map[string]config.Element),
		ElementMat: make(map[string]gocv.Mat),
		Sseract:    gosseract.NewClient(),
		sseractMu:  &sync.Mutex{},
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

func (a *ApiImpl) PressE(e config.Element, duration int) (bool, error) {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to press element[%s]: %w", e.Path, err)
	}
	p, val, err := a.FindE(e)
	if err != nil {
		return false, makeErr(err)
	}
	if val < e.Threshold {
		return false, nil
	}
	err = a.Press(p.X+e.Offset.X, p.Y+e.Offset.Y, duration)
	if err != nil {
		return false, makeErr(err)
	}
	return true, nil
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

func (h *SwipeHandler) Action(duration int) (image.Point, image.Point, bool, error) {
	if h.e1.Name != "" {
		p, val, err := h.api.FindE(h.e1)
		if err != nil {
			return image.ZP, image.ZP, false, err
		}
		if val < h.e1.Threshold {
			return image.ZP, image.ZP, false, nil
		}
		h.p1 = p
	}
	if h.e2.Name != "" {
		p, val, err := h.api.FindE(h.e2)
		if err != nil {
			return image.ZP, image.ZP, false, err
		}
		if val < h.e1.Threshold {
			return image.ZP, image.ZP, false, nil
		}
		h.p2 = p
	}
	err := h.api.Device.Input.Swipe(h.p1.X, h.p1.Y, h.p2.X, h.p2.Y, duration)
	if err != nil {
		return image.ZP, image.ZP, false, err
	}
	return h.p1, h.p2, true, nil
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
	if e.Img == "" {
		return image.ZP, 0, makeErr(errors.New("must be an \"Img\" field"))
	}
	tmpl := a.ElementMat[e.Path]
	img, err := a.ScreencapToMat()
	if err != nil {
		return image.ZP, 0, makeErr(err)
	}
	val, point, err := cv.Find(img, tmpl)
	if err != nil {
		return image.ZP, 0, makeErr(err)
	}
	return point, val, nil
}

func (a *ApiImpl) Screencap() ([]byte, error) {
	imgB, err := a.Device.ADB("shell /data/local/tmp/screencap 100")
	if err != nil {
		return nil, fmt.Errorf("failed to get screencap: %w", err)
	}
	return imgB, nil
}

func (a *ApiImpl) ScreencapToMat() (gocv.Mat, error) {
	if a.Img != nil {
		return *a.Img, nil
	}
	imgB, err := a.Screencap()
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
	img, err := a.ScreencapToMat()
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

func (a *ApiImpl) Adb(cmd string) ([]byte, error) {
	out, err := a.Device.ADB(cmd)
	if err != nil {
		return nil, fmt.Errorf("adb error: %w", err)
	}
	return out, nil
}

func (a *ApiImpl) Ocr(x1, y1, x2, y2 int) (string, error) {
	r := image.Rect(x1, y1, x2, y2)
	makeErr := func(err error) error {
		return fmt.Errorf("ocr error: %w", err)
	}
	imgB, err := a.Screencap()
	if err != nil {
		return "", makeErr(err)
	}
	buf := bytes.NewBuffer(imgB)
	img, _, err := image.Decode(buf)
	if err != nil {
		return "", makeErr(err)
	}
	subImg := img.(*image.YCbCr).SubImage(r)
	buf.Reset()
	err = jpeg.Encode(buf, subImg, &jpeg.Options{Quality: 100})
	if err != nil {
		return "", makeErr(err)
	}
	os.WriteFile("test.jpg", buf.Bytes(), 0660)
	a.sseractMu.Lock()
	defer a.sseractMu.Unlock()
	err = a.Sseract.SetImageFromBytes(buf.Bytes())
	if err != nil {
		return "", makeErr(err)
	}
	out, err := a.Sseract.Text()
	if err != nil {
		return "", makeErr(err)
	}
	return out, nil
}
