package engine

import (
	"io"
	"os"
	"path"
	"time"

	"github.com/HumXC/adb-helper"
	"github.com/HumXC/give-me-time/engine/config"
	"github.com/HumXC/give-me-time/engine/scripts"
	"golang.org/x/exp/slog"
)

type Client struct {
	Info    *config.Info
	Script  scripts.Script
	Device  adb.Device
	LogFile *os.File
}

func (c *Client) Start() error {
	err := c.Script.Run()
	if err != nil {
		return err
	}
	return nil
}
func LoadProject(projectPath string, device adb.Device) (*Client, error) {
	// 创建日志文件: log/*.log
	_ = os.Mkdir(path.Join(projectPath, "log"), 0755)
	logFile, err := os.OpenFile(path.Join(projectPath, "log", time.Now().Format(time.DateTime)+".log"),
		os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	logWriter := io.MultiWriter(logFile, os.Stdout)
	log := slog.New(LogHandler(logWriter))
	LogPrintHead(logWriter, projectPath, device)

	// info.json
	info, err := config.LoadInfo(path.Join(projectPath, "info.json"))
	if err != nil {
		return nil, err
	}
	LogPrintInfo(logWriter, *info)

	// element.json
	elm, err := config.LoadElement(path.Join(projectPath, "element.json"))
	if err != nil {
		return nil, err
	}
	LogPrintElement(logWriter, elm)

	// main.lua
	scr, err := scripts.LoadScript(device, *log, path.Join(projectPath, "main.lua"), info, elm)
	if err != nil {
		return nil, err
	}
	c := &Client{
		Info:    info,
		Script:  scr,
		Device:  device,
		LogFile: logFile,
	}
	return c, nil
}
