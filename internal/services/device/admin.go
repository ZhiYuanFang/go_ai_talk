package device

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/eventkit"
	"hello/internal/shared/assetpath"
	sharedtypes "hello/internal/shared/types"
	"hello/internal/services/workeroutbox"

	"github.com/gogf/gf/v2/frame/g"
)

const fixedDeviceAdminPassword = "a521521521"

var (
	ErrDeviceExists        = errors.New("设备号已存在")
	ErrDeviceNotRegistered = errors.New("设备未注册，请先注册设备号")
	ErrEventExists         = errors.New("事件已存在")
	ErrEventNotFound       = errors.New("事件不存在")
	ErrActionNotFound      = errors.New("动作不存在")
)

type service struct{}

var insAdmin = &service{}

// DeviceAdmin 返回设备管理领域服务实现。
func DeviceAdmin() *service { return insAdmin }

func (s *service) VerifyPassword(password string) bool { return strings.TrimSpace(password) == fixedDeviceAdminPassword }

func (s *service) Register(ctx context.Context, deviceNo string) (int64, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return 0, errors.New("deviceNo 不能为空")
	}
	count, err := dao.User.Ctx(ctx).Where(dao.User.Columns().DeviceNo, deviceNo).Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, ErrDeviceExists
	}
	activeTime := time.Now().Unix()
	_, err = dao.User.Ctx(ctx).Data(g.Map{dao.User.Columns().DeviceNo: deviceNo, dao.User.Columns().ActiveTime: activeTime}).Insert()
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique") {
			return 0, ErrDeviceExists
		}
		return 0, err
	}
	return activeTime, nil
}

func (s *service) EnsureRegistered(ctx context.Context, deviceNo string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return errors.New("deviceNo 不能为空")
	}
	count, err := dao.User.Ctx(ctx).Where(dao.User.Columns().DeviceNo, deviceNo).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrDeviceNotRegistered
	}
	return nil
}

// SaveUserProfile 更新 user 表画像字段；若开启 outbox 中继且配置了 WORKER_SERVICE_URL，则经 worker HTTP 写入 domain_outbox。
func (s *service) SaveUserProfile(ctx context.Context, deviceNo string, birthdayUnixSec int64, sex int) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return nil
	}
	if sex > 0 {
		sex = 1
	}
	_, err := dao.User.Ctx(ctx).Where(dao.User.Columns().DeviceNo, deviceNo).Data(g.Map{
		dao.User.Columns().Birthday: birthdayUnixSec,
		dao.User.Columns().Sex:      sex,
	}).Update()
	if err != nil {
		return err
	}
	if isDeviceOutboxRelayEnabled() {
		_ = enqueueUserProfileOutbox(ctx, deviceNo, birthdayUnixSec, sex)
	}
	_ = deviceCache.setUserProfile(ctx, cachedUserProfile{
		DeviceNo: deviceNo,
		Birthday: birthdayUnixSec,
		Sex:      sex,
	})
	return nil
}

// InsertVoiceActionRecord 语音链路写入动作表，名称已存在时返回错误。
func (s *service) InsertVoiceActionRecord(ctx context.Context, name, targetType string) error {
	name = strings.TrimSpace(name)
	targetType = strings.TrimSpace(targetType)
	if name == "" {
		return errors.New("动作名称不能为空")
	}
	actions, err := s.ListActionsForAdmin(ctx)
	if err != nil {
		return err
	}
	for _, action := range actions {
		if strings.EqualFold(strings.TrimSpace(action.Name), name) {
			return errors.New("动作名称已存在")
		}
	}
	_, err = dao.Action.Ctx(ctx).Insert(&entity.Action{Name: name, TargetType: targetType})
	if err != nil {
		return err
	}
	_ = enqueueDeviceProjectionEvent(ctx, eventkit.RoutingDeviceActionChanged, map[string]interface{}{
		"event_id":    fmt.Sprintf("device-action-changed-%d", time.Now().UnixNano()),
		"version":     time.Now().UnixNano(),
		"occurred_at": time.Now().Format(time.RFC3339Nano),
	})
	if rows, listErr := s.ListActions(ctx); listErr == nil {
		_ = deviceCache.setActionOptions(ctx, rows)
	}
	return nil
}

// InsertOrGetEventByNeedle 统一意图路径下按名称插入事件并回读。
func (s *service) InsertOrGetEventByNeedle(ctx context.Context, needle string, eventType string) (entity.Event, error) {
	needle = strings.TrimSpace(needle)
	if needle == "" {
		return entity.Event{}, errors.New("事件名为空")
	}
	created := entity.Event{Name: needle, EventType: NormalizeEventType(eventType)}
	if _, insErr := dao.Event.Ctx(ctx).Insert(&created); insErr != nil {
		// 并发或重复名称可能导致唯一约束失败，后续回读仍可拿到已有行。
		_ = insErr
	}
	_ = enqueueDeviceProjectionEvent(ctx, eventkit.RoutingDeviceEventChanged, map[string]interface{}{
		"event_id":    fmt.Sprintf("device-event-changed-%d", time.Now().UnixNano()),
		"version":     time.Now().UnixNano(),
		"occurred_at": time.Now().Format(time.RFC3339Nano),
	})
	refreshEventOptionsCacheAfterMutate(ctx)
	var inserted entity.Event
	err := dao.Event.Ctx(ctx).Where(dao.Event.Columns().Name, needle).OrderDesc(dao.Event.Columns().Id).Limit(1).Scan(&inserted)
	return inserted, err
}

// ApplyDeepSeekEventExtractPersistence DeepSeek 抽取结果落库：合并 extra_names 或插入新事件。
func (s *service) ApplyDeepSeekEventExtractPersistence(ctx context.Context, out entity.Event) (entity.Event, string, error) {
	name := strings.TrimSpace(out.Name)
	if name == "" {
		return entity.Event{}, "", errors.New("未抽取到事件名称")
	}
	out.Name = name
	eventList, listErr := s.ListEvents(ctx)
	if listErr != nil {
		return entity.Event{}, "", listErr
	}
	oldEvent := entity.Event{}
	for _, e := range eventList {
		if strings.EqualFold(strings.TrimSpace(e.Name), name) {
			oldEvent = e
			break
		}
	}
	targetName := strings.TrimSpace(out.ExtraNames)
	if oldEvent.Id > 0 {
		if strings.TrimSpace(out.ExtraNames) == "" {
			return out, out.Name, errors.New("事件名称已存在")
		}
		extraNames := strings.Split(oldEvent.ExtraNames, ",")
		for _, extraName := range extraNames {
			if strings.TrimSpace(extraName) == strings.TrimSpace(out.ExtraNames) {
				return out, strings.TrimSpace(out.ExtraNames), errors.New("事件名称已存在")
			}
		}
		merged := strings.TrimSpace(out.ExtraNames)
		if strings.TrimSpace(oldEvent.ExtraNames) != "" {
			merged = strings.Join(extraNames, ",") + "," + merged
		}
		out.ExtraNames = merged
		_, err := dao.Event.Ctx(ctx).Where(dao.Event.Columns().Name, name).Data(g.Map{
			dao.Event.Columns().ExtraNames: merged,
		}).Update()
		if err != nil {
			return entity.Event{}, "", err
		}
		_ = enqueueDeviceProjectionEvent(ctx, eventkit.RoutingDeviceEventChanged, map[string]interface{}{
			"event_id":    fmt.Sprintf("device-event-changed-%d", time.Now().UnixNano()),
			"version":     time.Now().UnixNano(),
			"occurred_at": time.Now().Format(time.RFC3339Nano),
		})
		refreshEventOptionsCacheAfterMutate(ctx)
		return out, targetName, nil
	}
	targetName = out.Name
	out.EventType = NormalizeEventType(out.EventType)
	if _, err := dao.Event.Ctx(ctx).Insert(&out); err != nil {
		return entity.Event{}, "", err
	}
	_ = enqueueDeviceProjectionEvent(ctx, eventkit.RoutingDeviceEventChanged, map[string]interface{}{
		"event_id":    fmt.Sprintf("device-event-changed-%d", time.Now().UnixNano()),
		"version":     time.Now().UnixNano(),
		"occurred_at": time.Now().Format(time.RFC3339Nano),
	})
	refreshEventOptionsCacheAfterMutate(ctx)
	return out, targetName, nil
}

func (s *service) UpdateLastTalk(ctx context.Context, deviceNo, ask, answer string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return errors.New("deviceNo 不能为空")
	}
	if err := s.EnsureRegistered(ctx, deviceNo); err != nil {
		return err
	}
	_, err := dao.User.Ctx(ctx).Where(dao.User.Columns().DeviceNo, deviceNo).Data(g.Map{
		dao.User.Columns().LastTalkTime:   time.Now().Unix(),
		dao.User.Columns().LastTalkAsk:    strings.TrimSpace(ask),
		dao.User.Columns().LastTalkAnswer: strings.TrimSpace(answer),
	}).Update()
	return err
}

func (s *service) List(ctx context.Context) ([]entity.User, error) {
	return listUsersWithRetry(ctx)
}

func eventListFields() []interface{} {
	c := dao.Event.Columns()
	return []interface{}{
		c.Id, c.Name, c.EventType, c.ExtraNames, c.Logo, c.Color,
	}
}

func normalizeEventRows(rows []entity.Event) {
	for i := range rows {
		rows[i].Logo = assetpath.Normalize(rows[i].Logo)
	}
}

func (s *service) AddEvent(ctx context.Context, name string, eventType string, extraNames, color, logoPath string) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("事件名称不能为空")
	}
	if err := ValidateEventColor(color); err != nil {
		return 0, err
	}
	if err := ValidateEventType(eventType); err != nil {
		return 0, err
	}
	eventType = NormalizeEventType(eventType)
	count, err := dao.Event.Ctx(ctx).Where(dao.Event.Columns().Name, name).Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, ErrEventExists
	}
	logoPath = assetpath.Normalize(strings.TrimSpace(logoPath))
	result, err := dao.Event.Ctx(ctx).Data(g.Map{
		dao.Event.Columns().Name:       name,
		dao.Event.Columns().EventType:  eventType,
		dao.Event.Columns().ExtraNames: strings.TrimSpace(extraNames),
		dao.Event.Columns().Color:        strings.TrimSpace(color),
		dao.Event.Columns().Logo:         logoPath,
	}).Insert()
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique") {
			return 0, ErrEventExists
		}
		return 0, err
	}
	id, _ := result.LastInsertId()
	if err == nil {
		_ = enqueueDeviceProjectionEvent(ctx, eventkit.RoutingDeviceEventChanged, map[string]interface{}{
			"event_id":    fmt.Sprintf("device-event-changed-%d", time.Now().UnixNano()),
			"version":     time.Now().UnixNano(),
			"occurred_at": time.Now().Format(time.RFC3339Nano),
		})
		// 事件元数据有变更时，从 DB 重建缓存快照（勿经 ListEvents，避免写回旧缓存）。
		refreshEventOptionsCacheAfterMutate(ctx)
	}
	return id, err
}

func (s *service) ListEvents(ctx context.Context) ([]entity.Event, error) {
	if rows, ok, err := deviceCache.getEventOptions(ctx); err == nil && ok {
		normalizeEventRows(rows)
		return rows, nil
	}
	rows := make([]entity.Event, 0)
	err := dao.Event.Ctx(ctx).Fields(eventListFields()...).OrderAsc(dao.Event.Columns().Id).Scan(&rows)
	if err == nil {
		normalizeEventRows(rows)
		_ = deviceCache.setEventOptions(ctx, rows)
	}
	return rows, err
}

func (s *service) UpdateEvent(ctx context.Context, id int64, name string, eventType string, extraNames, color, logoPath string) error {
	if id <= 0 {
		return errors.New("事件ID无效")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("事件名称不能为空")
	}
	if err := ValidateEventColor(color); err != nil {
		return err
	}
	if err := ValidateEventType(eventType); err != nil {
		return err
	}
	eventType = NormalizeEventType(eventType)
	idCount, err := dao.Event.Ctx(ctx).Where(dao.Event.Columns().Id, id).Count()
	if err != nil {
		return err
	}
	if idCount == 0 {
		return ErrEventNotFound
	}
	nameCount, err := dao.Event.Ctx(ctx).
		Where(dao.Event.Columns().Name, name).
		Where(fmt.Sprintf("%s<>?", dao.Event.Columns().Id), id).
		Count()
	if err != nil {
		return err
	}
	if nameCount > 0 {
		return ErrEventExists
	}
	data := g.Map{
		dao.Event.Columns().Name:       name,
		dao.Event.Columns().EventType:  eventType,
		dao.Event.Columns().ExtraNames: strings.TrimSpace(extraNames),
		dao.Event.Columns().Color:        strings.TrimSpace(color),
	}
	if lp := assetpath.Normalize(strings.TrimSpace(logoPath)); lp != "" {
		data[dao.Event.Columns().Logo] = lp
	}
	_, err = dao.Event.Ctx(ctx).Where(dao.Event.Columns().Id, id).Data(data).Update()
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique") {
			return ErrEventExists
		}
	}
	if err == nil {
		_ = enqueueDeviceProjectionEvent(ctx, eventkit.RoutingDeviceEventChanged, map[string]interface{}{
			"event_id":    fmt.Sprintf("device-event-changed-%d", time.Now().UnixNano()),
			"version":     time.Now().UnixNano(),
			"occurred_at": time.Now().Format(time.RFC3339Nano),
		})
		refreshEventOptionsCacheAfterMutate(ctx)
	}
	return err
}

func (s *service) DeleteEvent(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("事件ID无效")
	}
	n, err := dao.Event.Ctx(ctx).Where(dao.Event.Columns().Id, id).Count()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrEventNotFound
	}
	_, err = dao.Event.Ctx(ctx).Where(dao.Event.Columns().Id, id).Delete()
	if err == nil {
		_ = enqueueDeviceProjectionEvent(ctx, eventkit.RoutingDeviceEventChanged, map[string]interface{}{
			"event_id":    fmt.Sprintf("device-event-changed-%d", time.Now().UnixNano()),
			"version":     time.Now().UnixNano(),
			"occurred_at": time.Now().Format(time.RFC3339Nano),
		})
		refreshEventOptionsCacheAfterMutate(ctx)
	}
	return err
}

func (s *service) ListQA(ctx context.Context) ([]entity.Qa, error) {
	// qa 表仅在 voice 库存在，device 进程经 HTTP 委派到 voice-service（见 fetchQaListFromVoiceHTTP）。
	return fetchQaListFromVoiceHTTP(ctx)
}

func (s *service) ListActionsForAdmin(ctx context.Context) ([]sharedtypes.AdminActionItem, error) {
	actions, err := s.ListActions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]sharedtypes.AdminActionItem, 0, len(actions))
	for _, a := range actions {
		out = append(out, sharedtypes.AdminActionItem{
			Id:              a.Id,
			Name:            a.Name,
			TargetType:      a.TargetType,
			TargetTypeLabel: actionTargetTypeChinese(a.TargetType),
		})
	}
	return out, nil
}

// ListActions 返回动作列表（含缓存优先读）。
func (s *service) ListActions(ctx context.Context) ([]entity.Action, error) {
	if rows, ok, err := deviceCache.getActionOptions(ctx); err == nil && ok {
		return rows, nil
	}
	rows := make([]entity.Action, 0)
	err := dao.Action.Ctx(ctx).Fields(
		dao.Action.Columns().Id,
		dao.Action.Columns().Name,
		dao.Action.Columns().TargetType,
	).OrderAsc(dao.Action.Columns().Id).Scan(&rows)
	if err == nil {
		_ = deviceCache.setActionOptions(ctx, rows)
	}
	return rows, err
}

func (s *service) UpdateAction(ctx context.Context, id int64, name, targetType string) error {
	if id <= 0 {
		return errors.New("动作ID无效")
	}
	name = strings.TrimSpace(name)
	targetType = strings.TrimSpace(targetType)
	if name == "" {
		return errors.New("动作名称不能为空")
	}
	if targetType == "" {
		return errors.New("动作目标不能为空")
	}
	n, err := dao.Action.Ctx(ctx).Where(dao.Action.Columns().Id, id).Count()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrActionNotFound
	}
	_, err = dao.Action.Ctx(ctx).Where(dao.Action.Columns().Id, id).Data(g.Map{
		dao.Action.Columns().Name:       name,
		dao.Action.Columns().TargetType: targetType,
	}).Update()
	if err == nil {
		_ = enqueueDeviceProjectionEvent(ctx, eventkit.RoutingDeviceActionChanged, map[string]interface{}{
			"event_id":    fmt.Sprintf("device-action-changed-%d", time.Now().UnixNano()),
			"version":     time.Now().UnixNano(),
			"occurred_at": time.Now().Format(time.RFC3339Nano),
		})
		if rows, listErr := s.ListActions(ctx); listErr == nil {
			_ = deviceCache.setActionOptions(ctx, rows)
		}
	}
	return err
}

func (s *service) DeleteAction(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("动作ID无效")
	}
	n, err := dao.Action.Ctx(ctx).Where(dao.Action.Columns().Id, id).Count()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrActionNotFound
	}
	_, err = dao.Action.Ctx(ctx).Where(dao.Action.Columns().Id, id).Delete()
	if err == nil {
		_ = enqueueDeviceProjectionEvent(ctx, eventkit.RoutingDeviceActionChanged, map[string]interface{}{
			"event_id":    fmt.Sprintf("device-action-changed-%d", time.Now().UnixNano()),
			"version":     time.Now().UnixNano(),
			"occurred_at": time.Now().Format(time.RFC3339Nano),
		})
		if rows, listErr := s.ListActions(ctx); listErr == nil {
			_ = deviceCache.setActionOptions(ctx, rows)
		}
	}
	return err
}

func listUsersWithRetry(ctx context.Context) ([]entity.User, error) {
	const maxAttempts = 3
	backoff := 300 * time.Millisecond
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		rows := make([]entity.User, 0)
		err := dao.User.Ctx(ctx).Fields(
			dao.User.Columns().DeviceNo,
			dao.User.Columns().ActiveTime,
			dao.User.Columns().LastTalkTime,
			dao.User.Columns().LastTalkAsk,
			dao.User.Columns().LastTalkAnswer,
		).OrderAsc(dao.User.Columns().Id).Scan(&rows)
		if err == nil {
			return rows, nil
		}
		lastErr = err
		if !isRetryableDBErr(err) || attempt == maxAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 2
	}
	return nil, lastErr
}

func isRetryableDBErr(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	for _, s := range []string{"dial tcp", "connectex", "i/o timeout", "connection reset", "connection refused", "broken pipe"} {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

func actionTargetTypeChinese(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "start":
		return "开始"
	case "end":
		return "结束"
	case "one":
		return "单次"
	case "exit":
		return "退出"
	case "suggest":
		return "建议"
	case "search":
		return "查询"
	case "conversation":
		return "闲聊"
	default:
		return "未知"
	}
}

func enqueueDeviceProjectionEvent(ctx context.Context, routingKey eventkit.RouteKey, payload map[string]interface{}) error {
	return enqueueDomainOutboxToHistoryRelay(ctx, routingKey, payload)
}

func isDeviceOutboxRelayEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("OUTBOX_RELAY_ENABLED")))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func enqueueUserProfileOutbox(ctx context.Context, deviceNo string, birthdayUnixSec int64, sex int) error {
	return enqueueDomainOutboxToHistoryRelay(ctx, eventkit.RoutingDeviceUserProfileUpdated, map[string]interface{}{
		"event_id":    fmt.Sprintf("device-user-profile-updated-%d", time.Now().UnixNano()),
		"version":     time.Now().UnixNano(),
		"device_no":   deviceNo,
		"birthday":    birthdayUnixSec,
		"sex":         sex,
		"occurred_at": time.Now().Format(time.RFC3339Nano),
	})
}

// enqueueDomainOutboxToHistoryRelay 将领域事件经 worker HTTP 写入 worker 库 domain_outbox（依赖 WORKER_SERVICE_URL）。
func enqueueDomainOutboxToHistoryRelay(ctx context.Context, routingKey eventkit.RouteKey, payload map[string]interface{}) error {
	if !routingKey.IsValid() {
		return errors.New("invalid routing key")
	}
	return workeroutbox.EnqueueDomainOutbox(ctx, routingKey, payload)
}
