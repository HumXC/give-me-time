package project_test

import (
	"testing"

	"github.com/HumXC/give-me-time/engine/project"
)

func TestVerifyInfo(t *testing.T) {
	good := project.Info{
		Name: "test",
		Runtime: project.Runtime{
			Name: "go",
			Run:  "go rum",
		},
	}
	bad1 := project.Info{
		Name: "",
		Runtime: project.Runtime{
			Name: "go",
			Run:  "go rum",
		},
	}
	bad2 := project.Info{
		Name: "dd",
		Runtime: project.Runtime{
			Name: "",
			Run:  "go rum",
		},
	}
	bad3 := project.Info{
		Name: "ddds",
		Runtime: project.Runtime{
			Name: "ds",
			Run:  "",
		},
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
	err = project.VerifyInfo(bad2)
	if err == nil {
		t.Error("case [bad2] should be an error")
		return
	}
	err = project.VerifyInfo(bad3)
	if err == nil {
		t.Error("case [bad3] should be an error")
		return
	}
}
func TestLoadInfo(t *testing.T) {
	info, err := project.LoadInfo("info_test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if info.Runtime.Health == "" {
		t.Fatal("the runtime.health should not be empty. ")
	}
	if info.Runtime.Run == "" {
		t.Fatal("the runtime.run should not be empty. ")
	}
}
