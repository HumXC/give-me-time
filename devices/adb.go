package devices

import (
	"errors"

	"github.com/HumXC/adb-helper"
)

type ADB struct {
	server adb.Server
}

func (a *ADB) FirstDevice() (*Device, error) {
	ds, err := a.server.Devices()
	if err != nil {
		return nil, err
	}
	for _, d := range ds {
		return &Device{
			Input: d.Input,
			ADB:   d.Cmd,
		}, nil
	}
	return nil, errors.New("no device")
}
func (a *ADB) List() ([]string, error) {
	ds, err := a.server.Devices()
	if err != nil {
		return nil, err
	}
	result := make([]string, len(ds), len(ds))
	for _, d := range ds {
		result = append(result, d.ID)
	}
	return result, nil
}
func (a *ADB) GetDevice(id string) (*Device, error) {
	ds, err := a.server.Devices()
	if err != nil {
		return nil, err
	}
	if id != "" {
		for _, d := range ds {
			if d.ID != id {
				continue
			}
			return &Device{
				Input: d.Input,
				ADB:   d.Cmd,
			}, nil
		}
	}
	return nil, errors.New(id + " not found")
}
func NewADB(adbPath string) ADB {
	return ADB{
		server: adb.NewServer(adb.NewADBRunner(adbPath)),
	}
}
