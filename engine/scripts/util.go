package scripts

import (
	"reflect"

	"github.com/Shopify/go-lua"
)

func PushValue(l *lua.State, v any) {
	if v == nil {
		l.PushNil()
		return
	}
	switch reflect.TypeOf(v).Kind() {
	case reflect.Float64:
		l.PushNumber(v.(float64))
	case reflect.Bool:
		l.PushBoolean(v.(bool))
	case reflect.String:
		l.PushString(v.(string))
	case reflect.Map:
		PushMap(l, v.(map[string]any))
	case reflect.Slice:
		PushSlice(l, v.([]any))
	}
}

func PushMap(l *lua.State, m map[string]any) {
	l.NewTable()
	for k, v := range m {
		l.PushString(k)
		PushValue(l, v)
		l.SetTable(-3)
	}
}

func PushSlice(l *lua.State, s []any) {
	l.NewTable()
	for i, v := range s {
		l.PushNumber(float64(i))
		PushValue(l, v)
		l.SetTable(-3)
	}
}

// 判断一个 lua 的 table 是 slice 还是 map
// 将返回 reflect.Map 或 reflect.Slice
// 将 table 视为 map 的定义：
// 如果 table 中有字符串索引，则被视为 map，否则被视为 slice
func LuaTableTypeOf(val reflect.Value) reflect.Kind {
	hash := val.Elem().FieldByName("hash")
	it := hash.MapRange()
	for it.Next() {
		if it.Key().Elem().Kind() == reflect.String {
			return reflect.Map
		}
		break
	}
	return reflect.Slice
}

// 将 lua 的 table 转换成 go map
// table 中不使用字符串索引的元素会被删除
func LuaTableToMap(val reflect.Value) map[string]any {
	hash := val.Elem().FieldByName("hash")
	var m map[string]any = map[string]any{}

	it := hash.MapRange()
	for it.Next() {
		v := it.Value().Elem()
		key := it.Key().Elem().String()
		value := LuaValue(v)
		m[key] = value
	}
	return m
}

// 将 lua 中对应的数据类型转换成 go 的数据类型：
// number -> float64
// string -> string
// bool -> bool
// table -> map/slice 见 LuaTableTypeOf 函数
// 其他未列出的数据类型将返回 nil
func LuaValue(v reflect.Value) any {
	var value any
	switch v.Kind() {
	case reflect.Float64:
		value = v.Float()
	case reflect.String:
		value = v.String()
	case reflect.Bool:
		value = v.Bool()
	case reflect.Ptr:
		if LuaTableTypeOf(v) == reflect.Slice {
			value = LuaTableToSlice(v)
		} else {
			value = LuaTableToMap(v)
		}
	}
	return value
}

// 将 lua 的 table 转换成 go slice
// 元素中的空值会被删除，所以无法保证有序
func LuaTableToSlice(val reflect.Value) []any {
	array := val.Elem().FieldByName("array")
	hash := val.Elem().FieldByName("hash")
	s := make([]any, 0)
	it := hash.MapRange()
	for it.Next() {
		v := it.Value().Elem()
		value := LuaValue(v)
		s = append(s, value)
	}
	for i := 0; i < array.Len(); i++ {
		value := LuaValue(array.Index(i).Elem())
		// 此处判断 nil 是因为 "array" 中可能存在额外的，不在预期内的 nil 值
		if value != nil {
			s = append(s, value)
		}
	}
	return s
}
