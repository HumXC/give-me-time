package config_test

import (
	"testing"

	"github.com/HumXC/give-me-time/engine/config"
)

func TestVerifyOption(t *testing.T) {
	good := &config.Info{
		Name: "test",
	}
	bad1 := &config.Info{
		Name: "",
	}
	err := config.VerifyOption(good)
	if err != nil {
		t.Error(err)
		return
	}
	err = config.VerifyOption(bad1)
	if err == nil {
		t.Error("case [bad1] should be an error")
		return
	}
}
func TestLoadOption(t *testing.T) {
	_, err := config.LoadInfo("test.json")
	if err != nil {
		t.Error(err)
	}
}
