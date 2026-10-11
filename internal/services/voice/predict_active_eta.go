// 进行中计时 ETA：Redis 权威全量列表，供 App 同步与 Python internal 读取。
// 与 PredictImminentPending（下次发生 + 延时推送）分键；本模块 MUST NOT 发 MQ / 推送。
// 设计见 openspec/changes/predict-active-timing-eta。
package voice

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	deviceclient "hello/internal/clients/device"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	// predictActiveEtaSyncLockTTL 同宝宝同步短锁，对齐 imminent 秒级窗口。
	predictActiveEtaSyncLockTTL = 5 * time.Second
)

// predictActiveEtaCache 经 Observer 包装的 Redis 访问（沿用 Imminent 权威态模式）。
var predictActiveEtaCache = cachekit.Default()

// PredictActiveEtaItem Redis / App / internal GET 共用条目（字段名与 Flutter 对齐）。
type PredictActiveEtaItem struct {
	RootEventId int64  `json:"rootEventId"`
	LeafEventId int64  `json:"leafEventId"`
	HistoryId   string `json:"historyId,omitempty"`
	Title       string `json:"title,omitempty"`
	StartAt     int64  `json:"startAt"`
	ExpectEndAt int64  `json:"expectEndAt"`
}

// SyncPredictActiveEta 全量替换宝宝进行中 ETA 列表。
// Args: wxID 触发用户；deviceNo 宝宝；items 全量列表（可空=清空）。
// Returns: 清洗后写入条数、错误。
// Side Effects: Redis 写/删；SetNX 短锁；MUST NOT PublishDelayed / 推送。
func SyncPredictActiveEta(ctx context.Context, wxID int64, deviceNo string, items []PredictActiveEtaItem) (int, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if wxID <= 0 {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "缺少登录用户")
	}
	if deviceNo == "" {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	// 绑机校验：会话 wx 当前绑定须与请求 deviceNo 一致。
	bound, err := deviceclient.FetchDeviceNoByWxID(ctx, wxID)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(bound) != deviceNo {
		return 0, gerror.NewCode(gcode.CodeNotAuthorized, "deviceNo 与当前绑机不一致")
	}

	lockKey, err := cachekit.PredictActiveEtaSyncLockKey(deviceNo)
	if err != nil {
		return 0, err
	}
	got, lockErr := predictActiveEtaCache.SetNXEX(ctx, lockKey, "1", predictActiveEtaSyncLockTTL)
	if lockErr != nil {
		glog.Warningf(ctx, "[predict-active-eta] synclock err deviceNoLen=%d err=%v", len(deviceNo), lockErr)
	} else if !got {
		return 0, gerror.NewCode(gcode.CodeOperationFailed, "同步繁忙，请稍后重试")
	}
	defer func() { _ = predictActiveEtaCache.Del(ctx, lockKey) }()

	cleaned := cleanPredictActiveEtaItems(items)
	etaKey, err := cachekit.PredictActiveEtaKey(deviceNo)
	if err != nil {
		return 0, err
	}
	if len(cleaned) == 0 {
		if delErr := predictActiveEtaCache.Del(ctx, etaKey); delErr != nil {
			glog.Warningf(ctx, "[predict-active-eta] clear failed deviceNoLen=%d err=%v", len(deviceNo), delErr)
			return 0, gerror.WrapCode(gcode.CodeInternalError, delErr, "清空进行中 ETA 失败")
		}
		glog.Infof(ctx, "[predict-active-eta] sync_ok wxId=%d deviceNoLen=%d count=0 cleared=1", wxID, len(deviceNo))
		return 0, nil
	}

	raw, _ := json.Marshal(cleaned)
	if setErr := predictActiveEtaCache.Set(ctx, etaKey, string(raw)); setErr != nil {
		return 0, gerror.WrapCode(gcode.CodeInternalError, setErr, "写入进行中 ETA 失败")
	}
	glog.Infof(ctx, "[predict-active-eta] sync_ok wxId=%d deviceNoLen=%d count=%d", wxID, len(deviceNo), len(cleaned))
	return len(cleaned), nil
}

// LoadPredictActiveEta 读取宝宝当前进行中 ETA。
// Args: deviceNo 宝宝设备号。
// Returns: 列表（缺键/空/非法 JSON 视为空列表，不报错）；系统级 Redis 错误向上返回。
func LoadPredictActiveEta(ctx context.Context, deviceNo string) ([]PredictActiveEtaItem, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "deviceNo 不能为空")
	}
	key, err := cachekit.PredictActiveEtaKey(deviceNo)
	if err != nil {
		return nil, err
	}
	val, ok, err := predictActiveEtaCache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	// 缺键或空值：按契约返回空列表，不报错。
	if !ok || strings.TrimSpace(val) == "" {
		return []PredictActiveEtaItem{}, nil
	}
	var list []PredictActiveEtaItem
	if unmarshalErr := json.Unmarshal([]byte(val), &list); unmarshalErr != nil {
		glog.Warningf(ctx, "[predict-active-eta] corrupt json deviceNoLen=%d err=%v", len(deviceNo), unmarshalErr)
		return []PredictActiveEtaItem{}, nil
	}
	if list == nil {
		return []PredictActiveEtaItem{}, nil
	}
	return list, nil
}

// cleanPredictActiveEtaItems 按 D3 丢弃非法项；同 leafEventId 后写覆盖；title trim；historyId 空串省略。
func cleanPredictActiveEtaItems(items []PredictActiveEtaItem) []PredictActiveEtaItem {
	cleaned := make([]PredictActiveEtaItem, 0, len(items))
	seen := make(map[int64]int) // leafEventId -> index in cleaned
	for _, it := range items {
		if it.RootEventId <= 0 || it.LeafEventId <= 0 {
			continue
		}
		if it.StartAt <= 0 || it.ExpectEndAt <= 0 || it.ExpectEndAt < it.StartAt {
			continue
		}
		it.Title = strings.TrimSpace(it.Title)
		it.HistoryId = strings.TrimSpace(it.HistoryId)
		if idx, ok := seen[it.LeafEventId]; ok {
			cleaned[idx] = it
			continue
		}
		seen[it.LeafEventId] = len(cleaned)
		cleaned = append(cleaned, it)
	}
	return cleaned
}
