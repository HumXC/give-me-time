package main

import "testing"

func TestParseArg(t *testing.T) {
	test := [][]string{
		// pass
		{"", "21", "true"},
		{"", "1", "false"},
		{"", "100", "T"},
		{"", "21", "1"},
		{"", "21", "0"},
		{"", "21"},
		// fail: 参数个数不足，参数不合法
		{"", "21", ""},
		{"", ""},
		{"", "-100", "true"},
		{"", "0", "true"},
		{"", "1000", "true"},
		{"", "-se12", "true"},
		{"", "1", "dd"},
		{"", "0", "true"},
		{"", "100", "-1"},
		{"", "2", "-0"},
	}
	result := []bool{
		true, true, true, true, true, true, false, false, false, false, false, false, false, false, false, false,
	}
	for i := 0; i < len(test); i++ {
		_, _, err := ParseArg(test[i])
		if (err == nil && result[i] == false) ||
			(err != nil && result[i] == true) {
			t.Errorf("用例[%d]不符合预期", i)
		}
	}
}
