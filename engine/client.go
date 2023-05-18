package engine

import (
	"os"

	"github.com/HumXC/adb-helper"
	"github.com/HumXC/give-me-time/engine/config"
)

type Client struct {
	Info    *config.Info
	Device  adb.Device
	LogFile *os.File
}
