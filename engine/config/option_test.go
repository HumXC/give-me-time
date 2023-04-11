package config_test

import (
	"testing"

	"github.com/HumXC/give-me-time/engine/config"
)

func TestVerifyOption(t *testing.T) {
	good := []config.Option{{
		Name:    "test1",
		Type:    "string",
		Default: "事实上",
		Select:  []any{"duiduidui", ""},
	}, {
		Name: "test2",
		Type: "number",
		// 这里要填入浮点数，不能填整数。否则不会 pass
		// 因为 VerifyOption 是由 LoadOption 调用，
		// json 库会将 json 里的数字转成 float64，应该无伤大雅
		Default: 0.0,
	}, {
		Name:    "test3",
		Type:    "bool",
		Default: false,
	},
	}
	err := config.VerifyOption(good)
	if err != nil {
		t.Errorf("此处不应该有错误")
	}
	bads := [][]config.Option{{{
		Name:    "test1",
		Type:    "",
		Default: "事实上",
		Select:  []any{"duiduidui", ""},
	}}, {{
		Name:    "",
		Type:    "number",
		Default: 0,
	}}, {{
		Name:    "test3",
		Type:    "bool",
		Default: 0,
	}}, {{
		Name:    "test4",
		Type:    "string",
		Default: 0,
	}}, {{
		Name:    "test5",
		Type:    "number",
		Default: "",
		Select:  []any{""},
	}}, {{
		Name:    "test6",
		Type:    "bool",
		Default: []any{0, ""},
	}}, {{
		Name:    "test5",
		Type:    "number",
		Default: 0.0,
	}, {
		Name:    "test5",
		Type:    "number",
		Default: 0.0,
	}}}
	for _, v := range bads {
		err = config.VerifyOption(v)
		if err == nil {
			t.Errorf("此处应该有错误")
		}
	}
}

func TestParseOption(t *testing.T) {
	opts := []config.Option{
		{
			Name:    "name",
			Type:    "string",
			Default: "ss",
		}, {
			Name:    "num",
			Type:    "number",
			Default: 19.2,
		}, {
			Name:    "bool",
			Type:    "bool",
			Default: false,
		},
	}
	result := map[string]any{
		"bool": true, "name": "jack", "num": 19.2,
	}
	userOption := "user_option_test.json"
	m, _ := config.ParseOption(opts, userOption)
	for k, v := range result {
		if m[k] != v {
			t.Errorf("want: %v, got: %v", v, m[k])
		}
	}
}
