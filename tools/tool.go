package tools

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/HumXC/adb-helper"
)

//go:embed screencap/screencap
var screencap []byte

const AndroidTmpDir = "/data/local/tmp"

func pushToAndroidTmp(cmd adb.ADBRunner, fileName string) error {
	_, err := cmd(fmt.Sprintf("push %s %s", fileName, AndroidTmpDir))
	return err
}
func createTemp(data []byte) (string, error) {
	f, err := os.CreateTemp("", "")
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = f.Write(data)
	return f.Name(), err
}

func chmodX(cmd adb.ADBRunner, fileName string) error {
	_, err := cmd("chmod +x " + fileName)
	return err
}

func InitTools(device adb.Device) error {
	screencapFile, err := createTemp(screencap)
	defer os.Remove(screencapFile)

	err = pushToAndroidTmp(device.Cmd, screencapFile)
	if err != nil {
		return err
	}
	return nil
}
