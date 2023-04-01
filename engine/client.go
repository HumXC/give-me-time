package engine

import (
	"path"

	"github.com/HumXC/give-me-time/devices"
	"github.com/HumXC/give-me-time/engine/config"
	"github.com/HumXC/give-me-time/engine/scripts"
)

type Client struct {
	Option *config.Info
	Script scripts.Script
	Device devices.Device
}

func (c *Client) Start() error {
	err := c.Script.Run()
	if err != nil {
		return err
	}
	return nil
}
func LoadProject(projectPath string, device devices.Device) (*Client, error) {
	name := path.Base(projectPath)
	opt, err := config.LoadInfo(path.Join(projectPath, name+".json"))
	if err != nil {
		return nil, err
	}
	elm, err := config.LoadElement(path.Join(projectPath, name+".json"))
	if err != nil {
		return nil, err
	}
	api, err := scripts.NewApi(device, elm)
	if err != nil {
		return nil, err
	}
	scr := scripts.LoadScript(path.Join(projectPath, name+".lua"), opt, elm, api)

	c := &Client{
		Option: opt,
		Script: scr,
		Device: device,
	}
	return c, nil
}
