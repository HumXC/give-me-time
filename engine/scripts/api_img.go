package scripts

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"

	"github.com/HumXC/adb-helper"
	"github.com/HumXC/give-me-time/engine/config"
	"github.com/Shopify/go-lua"
	"gocv.io/x/gocv"
	"golang.org/x/exp/slog"
)

// “Element” 类型参数是指在 lua 中以 “click(E.main.start)” 的形式调用
type ApiImg interface {
	LuaApi
	// 查找元素
	FindE(e string) (image.Point, float32, error)
	// 返回范围内的文字识别结果
	Ocr(x1, y1, x2, y2 int) (string, error)
	OcrE(e string) (string, error)
	// 锁定与解锁当前 Find 函数的对象
	Lock() error
	Unlock() error
}
type apiImgImpl struct {
	imgHander   ImgHandler
	screencap   ScreencapTool
	nowImg      []byte
	elementMat  map[string]gocv.Mat
	elementArea map[string]config.ElArea
}

func (a *apiImgImpl) FindE(e string) (image.Point, float32, error) {
	imgB, err := a.GetScreen()
	if err != nil {
		return image.ZP, 0, fmt.Errorf("can not find element [%s]: %w", e, err)
	}
	img, err := gocv.IMDecode(imgB, gocv.IMReadUnchanged)
	if err != nil {
		return image.ZP, 0, fmt.Errorf("can not find element [%s]: %w", e, err)
	}
	tmpl, ok := a.elementMat[e]
	if !ok {
		return image.ZP, -1, fmt.Errorf("img element [%s] undefiend", e)
	}
	v, p, err := a.imgHander.Find(img, tmpl)
	if err != nil {
		return image.ZP, 0, fmt.Errorf("can not find element [%s]: %w", e, err)
	}
	return p, v, nil
}

func (a *apiImgImpl) Ocr(x1, y1, x2, y2 int) (string, error) {
	img, err := a.screencap.ToByte()
	if err != nil {
		return "", fmt.Errorf("can not ocr [%d, %d - %d, %d]: %w", x1, y1, x2, y2, err)
	}
	str, err := a.imgHander.Ocr(img)
	if err != nil {
		return "", fmt.Errorf("can not ocr [%d, %d - %d, %d]: %w", x1, y1, x2, y2, err)
	}
	return str, nil
}

func (a *apiImgImpl) OcrE(e string) (string, error) {
	area, ok := a.elementArea[e]
	if !ok {
		return "", fmt.Errorf("area element [%s] undefiend", e)
	}
	img, err := a.GetScreen()
	if err != nil {
		return "", fmt.Errorf("can not ocr element [%s]: %w", e, err)
	}
	r := image.Rect(area.P1.X, area.P1.Y, area.P2.X, area.P2.Y)
	buf := bytes.NewBuffer(img)
	im, _, err := image.Decode(buf)
	if err != nil {
		return "", fmt.Errorf("can not ocr element [%s]: %w", e, err)
	}
	subImg := im.(*image.YCbCr).SubImage(r)
	buf.Reset()
	err = jpeg.Encode(buf, subImg, &jpeg.Options{Quality: 100})
	if err != nil {
		return "", err
	}
	str, err := a.imgHander.Ocr(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("can not ocr element [%s]: %w", e, err)
	}
	return str, nil
}

func (a *apiImgImpl) GetScreen() ([]byte, error) {
	if a.nowImg != nil {
		return a.nowImg, nil
	}
	return a.screencap.ToByte()
}

func (a *apiImgImpl) Lock() error {
	if a.nowImg != nil {
		return errors.New("can not be locked repeatedly")
	}
	img, err := a.screencap.ToByte()
	if err != nil {
		return fmt.Errorf("failed to lock: %w", err)
	}
	a.nowImg = img
	return nil
}

func (a *apiImgImpl) Unlock() error {
	if a.nowImg == nil {
		return errors.New("can not be unlocked repeatedly")
	}
	a.nowImg = nil
	return nil
}

func (a *apiImgImpl) ToLuaFunc(log slog.Logger) map[string]lua.Function {
	m := make(map[string]lua.Function)
	// find(element) (x, y, maxVal)
	m["find"] = func(l *lua.State) int {
		args := NewArgsPicker(l)
		element, ok := args.Element(1)
		if !ok {
			PushErr(log, l, NewArgsErr("img element", l.ToValue(1)))
			return 0
		}
		p, v, err := a.FindE(element)
		if err != nil {
			PushErr(log, l, err)
			return 0
		}
		l.PushInteger(p.X)
		l.PushInteger(p.Y)
		l.PushNumber(float64(v))
		log.Info(fmt.Sprintf(
			"find element [%s] on (%d, %d), val: %f", element, p.X, p.Y, v))
		return 3
	}
	// ocr(x1, y1, x2, y2) string
	m["ocr"] = func(l *lua.State) int {
		args := NewArgsPicker(l)
		if e, ok := args.Element(1); ok {
			str, err := a.OcrE(e)
			if err != nil {
				PushErr(log, l, err)
				return 0
			}
			l.PushString(str)
			log.Info(fmt.Sprintf("ocr element [%s]: %s", e, str))
			return 1
		}
		points := [4]int{}
		for i := 0; i < 4; i++ {
			_i, ok := args.Int(i + 1)
			if !ok {
				PushErr(log, l, NewArgsErr("number", l.ToValue(i+1)))
				return 0
			}
			points[i] = _i
		}
		str, err := a.Ocr(points[0], points[1], points[2], points[3])
		if err != nil {
			PushErr(log, l, err)
			return 0
		}
		l.PushString(str)
		log.Info(fmt.Sprintf("ocr (%d, %d)-(%d, %d): %s", points[0], points[1], points[2], points[3], str))
		return 1
	}
	// lock()
	m["lock"] = func(l *lua.State) int {
		err := a.Lock()
		if err != nil {
			PushErr(log, l, err)
			return 0
		}
		log.Info("locked")
		return 0
	}
	// unlock()
	m["unlock"] = func(l *lua.State) int {
		err := a.Unlock()
		if err != nil {
			PushErr(log, l, err)
			return 0
		}
		log.Info("unlocked")
		return 0
	}
	return m
}

func NewApiImg(adbCmd adb.ADBRunner, elementImg map[string]config.ElImg, elementArea map[string]config.ElArea) (ApiImg, error) {
	a := apiImgImpl{
		elementMat: make(map[string]gocv.Mat),
		imgHander:  newImgHander(),
		screencap:  &screencapToolImpl{adbCmd: adbCmd},
	}
	for k, e := range elementImg {
		mat, err := gocv.IMDecode(e.Img, gocv.IMReadUnchanged)
		if err != nil {
			return nil, fmt.Errorf("failed to decode from [%s] to gocv.Mat: %w", k, err)
		}
		a.elementMat[k] = mat
	}
	return &a, nil
}
