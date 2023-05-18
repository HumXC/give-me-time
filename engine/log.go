package engine

import (
	"fmt"
	"io"

	"github.com/HumXC/adb-helper"
	"github.com/HumXC/give-me-time/engine/project"
	"golang.org/x/exp/slog"
)

func LogHandler(o io.Writer) slog.Handler {
	return slog.NewTextHandler(o)
}

func LogPrintHead(log io.Writer, projectPath string, dev adb.Device) {
	fmtMsg := `Project: %s
Device:
	ID: %s
	IsOnline: %t
	Product: %s
	Model: %s
	ADBPath: %s
`
	msg := fmt.Sprintf(fmtMsg,
		projectPath,
		dev.ID, dev.IsOnline, dev.Product, dev.Model, dev.ADBPath)
	_, _ = io.WriteString(log, msg)
}
func LogPrintInfo(log io.Writer, info project.Info) {
	fmtMsg := `Info:
	Name: %s
	Discripyion: %s
	Version: %s
`
	msg := fmt.Sprintf(fmtMsg,
		info.Name, info.Discription, info.Version)
	_, _ = io.WriteString(log, msg)
}

func LogPrintElement(log io.Writer, element []project.Element) {
	msg := "Element:\n" + treeElement("	", element)
	_, _ = io.WriteString(log, msg)
}

func treeElement(prefix string, es []project.Element) string {
	if len(es) == 0 {
		return ""
	}
	msg := ""
	for i, e := range es {
		pre := "│  " + prefix
		if i == len(es)-1 {
			pre = prefix + "   "
			msg += prefix + "└── " + e.Name + "\n"
		} else {
			msg += prefix + "├── " + e.Name + "\n"
		}
		sub := treeElement(pre, e.Element)
		msg += sub
	}
	return msg
}
