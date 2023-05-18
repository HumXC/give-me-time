package project_test

import (
	"testing"

	"github.com/HumXC/give-me-time/engine/project"
)

func TestVerifyInfo(t *testing.T) {
	good := project.Info{
		Name: "test",
	}
	bad1 := project.Info{
		Name: "",
	}
	err := project.VerifyInfo(good)
	if err != nil {
		t.Error(err)
		return
	}
	err = project.VerifyInfo(bad1)
	if err == nil {
		t.Error("case [bad1] should be an error")
		return
	}
}
func TestLoadInfo(t *testing.T) {
	_, err := project.LoadInfo("info_test.json")
	if err != nil {
		t.Error(err)
	}
}
