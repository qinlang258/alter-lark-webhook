package tools

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

func GetMapStr(data g.Map, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func GetMapByte(data g.Map, key string) []byte {
	if val, ok := data[key].([]byte); ok {
		return val
	}
	return nil
}

func GetMapInt64(data g.Map, key string) int64 {
	if val, ok := data[key].(int64); ok {
		return val
	}
	return 0
}

func GetMapInt(data g.Map, key string) int {
	if val, ok := data[key].(int); ok {
		return val
	}
	return 0
}

func GetMapInt32(data g.Map, key string) int32 {
	if val, ok := data[key].(int32); ok {
		return val
	}
	return 0
}

func GetMapTime(data g.Map, key string) *gtime.Time {
	if val, ok := data[key].(*gtime.Time); ok {
		return val
	}
	return nil
}

func GetGjsonjson(data g.Map, key string) *gjson.Json {
	if val, ok := data[key].(*gjson.Json); ok {
		return val
	}
	return nil
}
