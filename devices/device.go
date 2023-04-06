package devices

import (
	"github.com/HumXC/adb-helper"
)

type Input interface {
	Press(x, y, duration int) error
	Swipe(x1, y1, x2, y2, duration int) error
}
type Device struct {
	Input Input
	ADB   adb.ADBRunner
	Info  adb.Device
}
