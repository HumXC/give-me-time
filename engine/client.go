package engine

import (
	"path"
)

type Client struct {
	Option *Option
	Script Script
}

func Load(project string) (*Client, error) {
	name := path.Base(project)
	opt, err := LoadOption(path.Join(project, name+".json"))
	if err != nil {
		return nil, err
	}
	scr := LoadScript(path.Join(project, name+".lua"), opt)
	c := &Client{
		Option: opt,
		Script: scr,
	}
	return c, nil
}
