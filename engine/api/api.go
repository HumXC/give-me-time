package api

import (
	"errors"
	"fmt"
	"image"
	"sync"

	"github.com/HumXC/adb-helper"
	"github.com/HumXC/give-me-time/cv"
	"github.com/otiai10/gosseract/v2"
	"gocv.io/x/gocv"
)

var ErrArgs = errors.New("argument error")

func NewArgsErr(want, got any) error {
	return fmt.Errorf("%w: want [%v], got [%v]", ErrArgs, want, got)
}

type ImgHandler interface {
	// 模版匹配
	Find(img gocv.Mat, tmpl gocv.Mat) (float32, image.Point, error)
	// Sseract 识别文字
	Ocr(img []byte) (string, error)
}
type imgHanderImpl struct {
	Sseract   *gosseract.Client
	SseractMu *sync.Mutex
}

func (i *imgHanderImpl) Find(img gocv.Mat, tmpl gocv.Mat) (float32, image.Point, error) {
	v, p, err := cv.Find(img, tmpl)
	if err != nil {
		err = fmt.Errorf("cv error: %w", err)
	}
	return v, p, err
}

func (i *imgHanderImpl) Ocr(img []byte) (string, error) {
	i.SseractMu.Lock()
	defer i.SseractMu.Unlock()
	err := i.Sseract.SetImageFromBytes(img)
	if err != nil {
		return "", fmt.Errorf("sseract error: %w", err)
	}
	text, err := i.Sseract.Text()
	if err != nil {
		err = fmt.Errorf("sseract error: %w", err)
	}
	return text, err
}

func newImgHander() ImgHandler {
	return &imgHanderImpl{
		Sseract:   gosseract.NewClient(),
		SseractMu: &sync.Mutex{},
	}
}

type ScreencapTool interface {
	// 获取设备屏幕的截图并输出为 []byte
	ToByte() ([]byte, error)
}
type screencapToolImpl struct {
	adbCmd adb.ADBRunner
}

func (s *screencapToolImpl) ToByte() ([]byte, error) {
	data, err := s.adbCmd("shell /data/local/tmp/screencap 50")
	if err != nil {
		return nil, fmt.Errorf("adb error: %w", err)
	}
	return data, err
}

type InputHandler interface {
	// 按下一个元素或坐标，duration 单位是 ms。duration 为 0 时，将会自动把 duration 赋值为 100
	Press(x, y, duration int) error
	// 滑动
	Swipe(x, y int) InputHandlerSwipeTo
}

type InputHandlerSwipeTo interface {
	To(x, y int) InputHandlerSwipeAction
}
type InputHandlerSwipeAction interface {
	// 第一个返回值是开始滑动的点，第二个返回值是滑动结束的点，第三个返回值表示是否成功
	Action(duration int) (image.Point, image.Point, bool, error)
}
