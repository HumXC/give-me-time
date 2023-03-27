package devices

import (
	"fmt"
	"strings"

	"github.com/HumXC/adb-helper"
)

type Input interface {
	Press(x, y, duration int) error
	Swipe(x1, y1, x2, y2, duration int) error
}
type Device struct {
	Input Input
	ADB   adb.ADBRunner
}

func (d *Device) Screenshot() ([]byte, error) {
	b, err := d.ADB("shell screencap -p ")
	if err != nil {
		return nil, fmt.Errorf("failed to capture screenshot: %w", err)
	}
	return b, nil
}

func (d *Device) StartApp(packageName string) error {
	_, err := d.ADB("shell am start " + packageName)
	return err
}

func (d *Device) StopApp(packageName string) error {
	pkgName, _, _ := strings.Cut(packageName, "/")
	_, err := d.ADB("shell am force-stop " + pkgName)
	return err
}
