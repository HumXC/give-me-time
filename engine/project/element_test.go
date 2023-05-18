package project_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/HumXC/give-me-time/engine/project"
)

func TestLoadElement(t *testing.T) {
	_, err := project.LoadElement("element_test.json")
	if err != nil {
		t.Error(err)
	}
}

func TestVerifyElement(t *testing.T) {
	// 正确的 Element
	good1 := []project.Element{
		{Name: "test"},
		{Name: "test2",
			Element: []project.Element{
				{Name: "test4"},
				{Name: "test5"},
				{Name: "test6"},
				{Name: "test7"},
				{Name: "test8"},
			}},
		{Name: "test3"},
		// 有 Name 包含符号 '.'
		{Name: "test.a"},
		// 有 Name 包含符号 '-'
		{Name: "test-a"},
	}
	// 空的 Element
	good2 := []project.Element{}

	// 不正确的 Element
	bads := map[string][]project.Element{
		// 同一节点下有重复的 Name
		"bad1": {
			{Name: "test"},
			{Name: "test", Type: "img"},
		},
		// 有 Name 为空
		"bad2": {
			{Name: "", Element: []project.Element{
				{Name: "test"},
			}},
			{Name: "test2", Element: []project.Element{
				{Name: "test3"},
			}},
		},
		// Type 不符合要求
		"bad3": {
			{Name: "dd", Type: "dd"},
		},
		"bad4": {
			{Name: "dd", Type: "imgd"},
		},
	}

	err := project.VerifyElement("", good1)
	if err != nil {
		t.Error("case [good1] verify failed:", err)
		return
	}
	err = project.VerifyElement("", good2)
	if err != nil {
		t.Error("case [good2] verify failed:", err)
		return
	}
	for k, v := range bads {
		err = project.VerifyElement("", v)
		if err == nil {
			t.Error("case [" + k + "] should be an error, but not")
			return
		}
	}
}

func TestSetType(t *testing.T) {
	b, err := os.ReadFile("element_test.json")
	if err != nil {
		t.Fatal(err)
	}
	es := make([]project.Element, 0)
	ms := make([]map[string]any, 0)
	err = json.Unmarshal(b, &es)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(b, &ms)
	if err != nil {
		t.Fatal(err)
	}
	result := map[string]string{
		"main":            "img",
		"game":            "",
		"main.start":      "",
		"main.text":       "area",
		"main.text.input": "point",
	}
	es = project.SetType(es, ms)

	m := make(map[string]project.Element)
	project.FlatElement(m, "", es)
	for k, e := range m {
		if result[k] != e.Type {
			t.Errorf("[%s] want: [%s], got: [%s]", k, result[k], e.Type)
			return
		}
	}
}
