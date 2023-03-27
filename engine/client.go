package engine

import (
	"path"

	"github.com/HumXC/give-me-time/devices"
)

type Client struct {
	Option *Option
	Script Script
	Device devices.Device
}

func (c *Client) Start() error {
	err := c.Device.StartApp(c.Option.App)
	if err != nil {
		return err
	}
	err = c.Script.Run()
	if err != nil {
		return err
	}
	return c.Device.StopApp(c.Option.App)
}
func LoadProject(project string, device devices.Device) (*Client, error) {
	name := path.Base(project)
	opt, err := LoadOption(path.Join(project, name+".json"))
	if err != nil {
		return nil, err
	}
	elm, err := LoadElement(path.Join(project, name+".json"))
	if err != nil {
		return nil, err
	}
	api, err := NewApi(device.Input, elm)
	if err != nil {
		return nil, err
	}
	scr := LoadScript(path.Join(project, name+".lua"), opt, elm, api)

	c := &Client{
		Option: opt,
		Script: scr,
		Device: device,
	}
	return c, nil
}
