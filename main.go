package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/HumXC/adb-helper"
	"github.com/HumXC/give-me-time/devices"
	"github.com/HumXC/give-me-time/engine"
)

var (
	deviceID    string
	projectName string
	adbPath     string
)

func init() {
	flag.StringVar(&deviceID, "device", "", "指定一个设备，如果为空，则使用第一个设备（如果有）")
	flag.StringVar(&projectName, "run", "", "指定一个工程")
	flag.StringVar(&adbPath, "adb", "adb", "指定 adb 的路径，默认值为 “adb”")
	flag.Parse()
}
func main() {
	if projectName == "" {
		os.Exit(0)
	}
	_adb := devices.NewADB(adbPath)
	var device *adb.Device
	if deviceID != "" {
		d, err := _adb.GetDevice(deviceID)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		device = d
	} else {
		d, err := _adb.FirstDevice()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		device = d
	}
	client, err := engine.LoadProject(projectName, *device)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = client.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
