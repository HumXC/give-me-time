package config_test

import (
	"testing"

	"github.com/HumXC/give-me-time/engine/config"
)

func TestVerifyInfo(t *testing.T) {
	good := &config.Info{
		Name: "test",
	}
	bad1 := &config.Info{
		Name: "",
	}
	err := config.VerifyInfo(good)
	if err != nil {
		t.Error(err)
		return
	}
	err = config.VerifyInfo(bad1)
	if err == nil {
		t.Error("case [bad1] should be an error")
		return
	}
}
func TestLoadInfo(t *testing.T) {
	_, err := config.LoadInfo("info_test.json")
	if err != nil {
		t.Error(err)
	}
}
