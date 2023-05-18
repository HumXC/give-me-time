package engine_test

import (
	"testing"

	"github.com/HumXC/adb-helper"
)

const TestProject = "../test/project-setting"

func TestLoad(t *testing.T) {
	server := adb.DefaultServer()
	_, err := server.Devices()
	if err != nil {
		t.Fatal(err)
	}

}
