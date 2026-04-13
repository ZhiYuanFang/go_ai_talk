package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

const fixedDeviceAdminPassword = "a521521521"

var (
	ErrDeviceExists        = errors.New("设备号已存在")
	ErrDeviceNotRegistered = errors.New("设备未注册，请先注册设备号")
	ErrEventExists         = errors.New("事件已存在")
	ErrEventNotFound       = errors.New("事件不存在")
	ErrIntentionNotFound   = errors.New("意图不存在")
	userCountByDevice      = func(ctx context.Context, deviceNo string) (int, error) {
		return dao.User.Ctx(ctx).Where(dao.User.Columns().DeviceNo, deviceNo).Count()
	}
	userInsertDevice = func(ctx context.Context, deviceNo, activeTime string) error {
		_, err := dao.User.Ctx(ctx).Data(g.Map{
			dao.User.Columns().DeviceNo:   deviceNo,
			dao.User.Columns().ActiveTime: activeTime,
		}).Insert()
		return err
	}
	userUpdateLastTalk = func(ctx context.Context, deviceNo, lastTalkTime, ask, answer string) error {
		_, err := dao.User.Ctx(ctx).Where(dao.User.Columns().DeviceNo, deviceNo).Data(g.Map{
			dao.User.Columns().LastTalkTime:   lastTalkTime,
			dao.User.Columns().LastTalkAsk:    ask,
			dao.User.Columns().LastTalkAnswer: answer,
		}).Update()
		return err
	}
	userListAll = func(ctx context.Context) ([]entity.User, error) {
		var rows []entity.User
		err := dao.User.Ctx(ctx).
			Fields(
				dao.User.Columns().DeviceNo,
				dao.User.Columns().ActiveTime,
				dao.User.Columns().LastTalkTime,
				dao.User.Columns().LastTalkAsk,
				dao.User.Columns().LastTalkAnswer,
			).
			OrderAsc(dao.User.Columns().Id).
			Scan(&rows)
		return rows, err
	}
	eventCountByName = func(ctx context.Context, name string) (int, error) {
		return dao.Event.Ctx(ctx).Where(dao.Event.Columns().Name, name).Count()
	}
	eventInsert = func(ctx context.Context, name string, needTime, needQuantity int) error {
		_, err := dao.Event.Ctx(ctx).Data(g.Map{
			dao.Event.Columns().Name:         name,
			dao.Event.Columns().NeedTime:     needTime,
			dao.Event.Columns().NeedQuantity: needQuantity,
		}).Insert()
		return err
	}
	eventCountByID = func(ctx context.Context, id int64) (int, error) {
		return dao.Event.Ctx(ctx).Where(dao.Event.Columns().Id, id).Count()
	}
	eventCountByNameExcludeID = func(ctx context.Context, id int64, name string) (int, error) {
		return dao.Event.Ctx(ctx).
			Where(dao.Event.Columns().Name, name).
			Where(fmt.Sprintf("%s<>?", dao.Event.Columns().Id), id).
			Count()
	}
	eventUpdateByID = func(ctx context.Context, id int64, name string, needTime, needQuantity int) error {
		_, err := dao.Event.Ctx(ctx).
			Where(dao.Event.Columns().Id, id).
			Data(g.Map{
				dao.Event.Columns().Name:         name,
				dao.Event.Columns().NeedTime:     needTime,
				dao.Event.Columns().NeedQuantity: needQuantity,
			}).
			Update()
		return err
	}
	eventListAll = func(ctx context.Context) ([]entity.Event, error) {
		var rows []entity.Event
		err := dao.Event.Ctx(ctx).
			Fields(dao.Event.Columns().Id, dao.Event.Columns().Name, dao.Event.Columns().NeedTime, dao.Event.Columns().NeedQuantity).
			OrderAsc(dao.Event.Columns().Id).
			Scan(&rows)
		return rows, err
	}
	intentionCountByID = func(ctx context.Context, id int64) (int, error) {
		return dao.Intention.Ctx(ctx).Where(dao.Intention.Columns().Id, id).Count()
	}
	intentionUpdateUpperLimitByID = func(ctx context.Context, id int64, upperLimit int) error {
		_, err := dao.Intention.Ctx(ctx).
			Where(dao.Intention.Columns().Id, id).
			Data(g.Map{dao.Intention.Columns().UpperLimit: upperLimit}).
			Update()
		return err
	}
	intentionListAll = func(ctx context.Context) ([]entity.Intention, error) {
		var rows []entity.Intention
		err := dao.Intention.Ctx(ctx).
			Fields(dao.Intention.Columns().Id, dao.Intention.Columns().Name, dao.Intention.Columns().UpperLimit).
			OrderAsc(dao.Intention.Columns().Id).
			Scan(&rows)
		return rows, err
	}
)

type sDeviceAdmin struct{}

var insDeviceAdmin = sDeviceAdmin{}

func DeviceAdmin() DeviceAdminContract {
	return &insDeviceAdmin
}

var _ DeviceAdminContract = (*sDeviceAdmin)(nil)

func (s *sDeviceAdmin) VerifyPassword(password string) bool {
	return strings.TrimSpace(password) == fixedDeviceAdminPassword
}

func (s *sDeviceAdmin) Register(ctx context.Context, deviceNo string) (string, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return "", errors.New("deviceNo 不能为空")
	}

	count, err := userCountByDevice(ctx, deviceNo)
	if err != nil {
		return "", err
	}
	if count > 0 {
		return "", ErrDeviceExists
	}

	activeTime := nowText()
	err = userInsertDevice(ctx, deviceNo, activeTime)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			return "", ErrDeviceExists
		}
		return "", err
	}

	return activeTime, nil
}

func (s *sDeviceAdmin) EnsureRegistered(ctx context.Context, deviceNo string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return errors.New("deviceNo 不能为空")
	}

	count, err := userCountByDevice(ctx, deviceNo)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrDeviceNotRegistered
	}
	return nil
}

func (s *sDeviceAdmin) UpdateLastTalk(ctx context.Context, deviceNo, ask, answer string) error {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return errors.New("deviceNo 不能为空")
	}
	if err := s.EnsureRegistered(ctx, deviceNo); err != nil {
		return err
	}

	return userUpdateLastTalk(ctx, deviceNo, nowText(), strings.TrimSpace(ask), strings.TrimSpace(answer))
}

func (s *sDeviceAdmin) List(ctx context.Context) ([]DeviceAdminItem, error) {
	rows, err := listUsersWithRetry(ctx)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func listUsersWithRetry(ctx context.Context) ([]entity.User, error) {
	const maxAttempts = 3
	backoff := 300 * time.Millisecond
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		rows, err := userListAll(ctx)
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
	if errors.As(err, &netErr) {
		if netErr.Timeout() || netErr.Temporary() {
			return true
		}
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	for _, s := range []string{
		"dial tcp",
		"connectex",
		"i/o timeout",
		"connection reset",
		"connection refused",
		"broken pipe",
	} {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

func (s *sDeviceAdmin) AddEvent(ctx context.Context, name string, needTime, needQuantity int) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("事件名称不能为空")
	}
	if needTime > 0 {
		needTime = 1
	}
	if needQuantity > 0 {
		needQuantity = 1
	}

	count, err := eventCountByName(ctx, name)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrEventExists
	}

	err = eventInsert(ctx, name, needTime, needQuantity)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			return ErrEventExists
		}
		return err
	}
	return nil
}

func (s *sDeviceAdmin) ListEvents(ctx context.Context) ([]DeviceEventItem, error) {
	rows, err := eventListAll(ctx)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *sDeviceAdmin) UpdateEvent(ctx context.Context, id int64, name string, needTime, needQuantity int) error {
	if id <= 0 {
		return errors.New("事件ID无效")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("事件名称不能为空")
	}
	if needTime > 0 {
		needTime = 1
	}
	if needQuantity > 0 {
		needQuantity = 1
	}

	idCount, err := eventCountByID(ctx, id)
	if err != nil {
		return err
	}
	if idCount == 0 {
		return ErrEventNotFound
	}

	nameCount, err := eventCountByNameExcludeID(ctx, id, name)
	if err != nil {
		return err
	}
	if nameCount > 0 {
		return ErrEventExists
	}

	err = eventUpdateByID(ctx, id, name, needTime, needQuantity)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			return ErrEventExists
		}
		return err
	}
	return nil
}

func (s *sDeviceAdmin) ListIntentions(ctx context.Context) ([]entity.Intention, error) {
	return intentionListAll(ctx)
}

func (s *sDeviceAdmin) UpdateIntentionUpperLimit(ctx context.Context, id int64, upperLimit int) error {
	if id <= 0 {
		return errors.New("意图ID无效")
	}
	if upperLimit < 0 {
		return errors.New("upperLimit 不能小于0")
	}
	count, err := intentionCountByID(ctx, id)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrIntentionNotFound
	}
	return intentionUpdateUpperLimitByID(ctx, id, upperLimit)
}

func nowText() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
