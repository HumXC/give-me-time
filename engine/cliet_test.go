package engine_test

import (
	"testing"

	"github.com/HumXC/adb-helper"
	"github.com/HumXC/give-me-time/engine"
)

const TestProject = "../test/project-setting"

func TestLoad(t *testing.T) {
	server := adb.DefaultServer()
	ds, err := server.Devices()
	if err != nil {
		t.Fatal(err)
	}

	device := ds[0]

	client, err := engine.LoadProject("test/明日方舟", device)
	if err != nil {
		t.Fatal(err)
	}
	err = client.Start()
	if err != nil {
		t.Fatal(err)
	}
}
