package cachekit

import (
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/container/gvar"
)

// parseHashGetAllResult 解析 Redis HGETALL 结果。GoFrame adapter 对 HGETALL 的 []interface{}
// 会转为 flat []string，须优先用 (*gvar.Var).MapStrStr()（见 gf redis_conn.resultToVar）。
//
//	把 Redis 返回的“各种奇葩 Hash 结果”，统一转成 map[string]string
func parseHashGetAllResult(v interface{}) (map[string]string, error) {
	out := make(map[string]string)
	if v == nil {
		return out, nil
	}
	if vv, ok := v.(*gvar.Var); ok && vv != nil {
		if m := vv.MapStrStr(); len(m) > 0 {
			return m, nil
		}
		v = vv.Val()
	}
	if m, ok := v.(map[string]interface{}); ok {
		for k, val := range m {
			out[k] = hashFieldString(val)
		}
		return out, nil
	}
	if m, ok := v.(map[string]string); ok {
		return m, nil
	}
	if arr, ok := v.([]string); ok {
		for i := 0; i+1 < len(arr); i += 2 {
			k := strings.TrimSpace(arr[i])
			if k != "" {
				out[k] = arr[i+1]
			}
		}
		return out, nil
	}
	arr, ok := v.([]interface{})
	if !ok || len(arr) == 0 {
		return out, nil
	}
	for i := 0; i+1 < len(arr); i += 2 {
		k := hashFieldString(arr[i])
		val := hashFieldString(arr[i+1])
		if k != "" {
			out[k] = val
		}
	}
	return out, nil
}

func hashFieldString(v interface{}) string {
	if vv, ok := v.(*gvar.Var); ok && vv != nil {
		return hashFieldString(vv.Val())
	}
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return fmt.Sprint(v)
	}
}
