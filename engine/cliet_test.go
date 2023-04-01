package engine_test

import (
	"testing"

	"github.com/HumXC/adb-helper"
	"github.com/HumXC/give-me-time/devices"
	"github.com/HumXC/give-me-time/engine"
)

const TestProject = "../test/project-setting"

func TestLoad(t *testing.T) {
	server := adb.DefaultServer()
	ds, err := server.Devices()
	if err != nil {
		t.Fatal(err)
	}

	device := devices.Device{
		ADB:   ds[0].Cmd,
		Input: ds[0].Input,
	}

	client, err := engine.LoadProject(TestProject, device)
	if err != nil {
		t.Fatal(err)
	}
	err = client.Start()
	if err != nil {
		t.Fatal(err)
	}
}
