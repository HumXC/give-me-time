package engine

import (
	"os"

	"github.com/HumXC/adb-helper"
	"github.com/HumXC/give-me-time/engine/project"
)

type Client struct {
	Info    *project.Info
	Device  adb.Device
	LogFile *os.File
}
