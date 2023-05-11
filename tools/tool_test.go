package tools_test

import (
	"testing"

	"github.com/HumXC/adb-helper"
	"github.com/HumXC/give-me-time/tools"
)

func TestXxx(t *testing.T) {
	server := adb.DefaultServer()
	ds, err := server.Devices()
	if err != nil {
		t.Fatal(err)
	}
	err = tools.InitTools(ds[0])
	if err != nil {
		panic(err)
	}

}
