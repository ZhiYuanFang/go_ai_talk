package cachekit

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/container/gvar"
)

// parseZRevRangeWithScoresResult 解析 Redis ZREVRANGE ... WITHSCORES 的 Do 返回值。
// go-redis 9.x 经 GoFrame adapter 常返回嵌套 [[member, score], ...]；部分环境为扁平 [m,s,m,s,...]。
func parseZRevRangeWithScoresResult(v interface{}) []ZSetMemberScore {
	arr := gvar.New(v).Array()
	if len(arr) == 0 {
		return nil
	}
	// 首元素为二元组 → 嵌套形态（Feed MONITOR 误解析时 GEOPOS 会收到 "[\"id\",score]" 串）
	if pair := gvar.New(arr[0]).Array(); len(pair) >= 2 {
		out := make([]ZSetMemberScore, 0, len(arr))
		for _, item := range arr {
			row := gvar.New(item).Array()
			if len(row) < 2 {
				continue
			}
			member := zsetFieldString(row[0])
			if member == "" {
				continue
			}
			score, _ := strconv.ParseFloat(zsetFieldString(row[1]), 64)
			out = append(out, ZSetMemberScore{Member: member, Score: score})
		}
		return out
	}
	out := make([]ZSetMemberScore, 0, len(arr)/2)
	for i := 0; i+1 < len(arr); i += 2 {
		member := zsetFieldString(arr[i])
		if member == "" {
			continue
		}
		score, _ := strconv.ParseFloat(zsetFieldString(arr[i+1]), 64)
		out = append(out, ZSetMemberScore{Member: member, Score: score})
	}
	return out
}

func zsetFieldString(v interface{}) string {
	if vv, ok := v.(*gvar.Var); ok && vv != nil {
		return zsetFieldString(vv.Val())
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case []byte:
		return strings.TrimSpace(string(t))
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}
