package engine_test

import (
	"testing"

	"github.com/HumXC/give-me-time/engine"
)

func TestVerifyElement(t *testing.T) {
	// 正确的 Element
	good1 := []engine.Element{
		{Name: "test"},
		{Name: "test2",
			Element: []engine.Element{
				{Name: "test4"},
				{Name: "test5"},
				{Name: "test6"},
				{Name: "test7"},
				{Name: "test8"},
			}},
		{Name: "test3"},
	}
	// 空的 Element
	good2 := []engine.Element{}

	// 不正确的 Element
	// 同一节点下有重复的 Name
	bad1 := []engine.Element{
		{Name: "test"},
		{Name: "test"},
	}
	// 有 Name 为空
	bad2 := []engine.Element{
		{Name: "", Element: []engine.Element{
			{Name: "test"},
		}},
		{Name: "test2", Element: []engine.Element{
			{Name: "test3"},
		}},
	}
	// 有 Name 包含符号 '.'
	bad3 := []engine.Element{
		{Name: "test.a"},
	}
	err := engine.VerifyElement("", good1)
	if err != nil {
		t.Error("case [good1] verify failed:", err)
		return
	}
	err = engine.VerifyElement("", good2)
	if err != nil {
		t.Error("case [good2] verify failed:", err)
		return
	}
	err = engine.VerifyElement("", bad1)
	if err == nil {
		t.Error("case [bad1] should be an error, but not")
		return
	}
	err = engine.VerifyElement("", bad2)
	if err == nil {
		t.Error("case [bad2] should be an error, but not")
		return
	}
	err = engine.VerifyElement("", bad3)
	if err == nil {
		t.Error("case [bad3] should be an error, but not")
		return
	}
}
func TestVerifyOption(t *testing.T) {
	good := &engine.Option{
		Name: "test",
		App:  "test",
	}
	bad1 := &engine.Option{
		Name: "",
	}
	bad2 := &engine.Option{
		Name: "s",
		App:  "",
	}
	err := engine.VerifyOption(good)
	if err != nil {
		t.Error(err)
		return
	}
	err = engine.VerifyOption(bad1)
	if err == nil {
		t.Error("case [bad1] should be an error")
		return
	}
	err = engine.VerifyOption(bad2)
	if err == nil {
		t.Error("case [bad2] should be an error")
		return
	}
}
func TestLoadOption(t *testing.T) {
	_, err := engine.LoadOption("test.json")
	if err != nil {
		t.Error(err)
	}
}

func TestLoadElement(t *testing.T) {
	_, err := engine.LoadElement("test.json")
	if err != nil {
		t.Error(err)
	}
}
