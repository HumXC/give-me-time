package api

import (
	"fmt"
	"image"

	"github.com/HumXC/adb-helper"
)

type ApiAdb interface {
	InputHandler
	// 执行 adb 命令
	Cmd(string) ([]byte, error)
}

type apiAdbImpl struct {
	input adb.Input
	cmd   adb.ADBRunner
}

func (a *apiAdbImpl) Cmd(cmd string) ([]byte, error) {
	out, err := a.cmd(cmd)
	if err != nil {
		return nil, fmt.Errorf("adb error: %w", err)
	}
	return out, nil
}

func (a *apiAdbImpl) Press(x, y, duration int) error {
	return a.input.Press(x, y, duration)
}

type SwipeHandler struct {
	swipe  func(x1, y1, x2, y2, duration int) error
	p1, p2 image.Point
}

func (h *SwipeHandler) To(x, y int) InputHandlerSwipeAction {
	h.p2.X = x
	h.p2.Y = y
	return h
}

func (h *SwipeHandler) Action(duration int) (image.Point, image.Point, bool, error) {

	err := h.swipe(h.p1.X, h.p1.Y, h.p2.X, h.p2.Y, duration)
	if err != nil {
		return image.ZP, image.ZP, false, err
	}
	return h.p1, h.p2, true, nil
}

func (a *apiAdbImpl) Swipe(x, y int) InputHandlerSwipeTo {
	return &SwipeHandler{
		swipe: a.input.Swipe,
		p1:    image.Point{X: x, Y: y},
	}
}

func NewApiAdb(device adb.Device) ApiAdb {
	return &apiAdbImpl{
		input: device.Input,
		cmd:   device.Cmd,
	}
}
