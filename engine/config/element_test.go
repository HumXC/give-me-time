package config_test

import (
	"testing"

	"github.com/HumXC/give-me-time/engine/config"
)

func TestLoadElement(t *testing.T) {
	_, err := config.LoadElement("element_test.json")
	if err != nil {
		t.Error(err)
	}
}

func TestVerifyElement(t *testing.T) {
	// 正确的 Element
	good1 := []config.Element{
		{Name: "test"},
		{Name: "test2",
			Element: []config.Element{
				{Name: "test4"},
				{Name: "test5"},
				{Name: "test6"},
				{Name: "test7"},
				{Name: "test8"},
			}},
		{Name: "test3"},
	}
	// 空的 Element
	good2 := []config.Element{}

	// 不正确的 Element
	bads := map[string][]config.Element{
		// 同一节点下有重复的 Name
		"bad1": {
			{Name: "test"},
			{Name: "test"},
		},
		// 有 Name 为空
		"bad2": {
			{Name: "", Element: []config.Element{
				{Name: "test"},
			}},
			{Name: "test2", Element: []config.Element{
				{Name: "test3"},
			}},
		},
		// 有 Name 包含符号 '.'
		"bad3": {
			{Name: "test.a"},
		},
		// 有 Name 包含符号 '-'
		"bad4": {
			{Name: "test-a"},
		},
	}

	err := config.VerifyElement("", good1)
	if err != nil {
		t.Error("case [good1] verify failed:", err)
		return
	}
	err = config.VerifyElement("", good2)
	if err != nil {
		t.Error("case [good2] verify failed:", err)
		return
	}
	for k, v := range bads {
		err = config.VerifyElement("", v)
		if err == nil {
			t.Error("case [" + k + "] should be an error, but not")
			return
		}
	}
}
