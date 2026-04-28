package controller

import (
	"context"
	"testing"

	v1 "hello/api/v1"
	"hello/internal/model/entity"
	"hello/internal/service"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

type mockHistoryService struct {
	listHistoryFn   func(ctx context.Context, deviceNo string) ([]entity.History, error)
	deleteSuggestFn func(ctx context.Context, id int64, deviceNo string) error
}

func (m *mockHistoryService) ListHistory(ctx context.Context, deviceNo string) ([]entity.History, error) {
	if m.listHistoryFn != nil {
		return m.listHistoryFn(ctx, deviceNo)
	}
	return nil, nil
}

func (m *mockHistoryService) ListSuggest(ctx context.Context, deviceNo string) ([]entity.Suggest, error) {
	return nil, nil
}

func (m *mockHistoryService) DeleteSuggest(ctx context.Context, id int64, deviceNo string) error {
	if m.deleteSuggestFn != nil {
		return m.deleteSuggestFn(ctx, id, deviceNo)
	}
	return nil
}

func (m *mockHistoryService) ListEventOptions(ctx context.Context) ([]entity.Event, error) {
	return nil, nil
}

func (m *mockHistoryService) GetBirthday(ctx context.Context, deviceNo string) (string, int, error) {
	return "", 0, nil
}

func (m *mockHistoryService) SaveBirthday(ctx context.Context, deviceNo, birthday string, sex int) error {
	return nil
}

func (m *mockHistoryService) AddHistory(ctx context.Context, item entity.History) (int64, error) {
	return 0, nil
}

func (m *mockHistoryService) UpdateHistory(ctx context.Context, item entity.History) error {
	return nil
}

func (m *mockHistoryService) DeleteHistory(ctx context.Context, id int64, deviceNo string) error {
	return nil
}

var _ service.DeviceHistoryContract = (*mockHistoryService)(nil)

func TestHistoryCtrlListRequiresDeviceNo(t *testing.T) {
	ctrl := NewHistoryCtrl(&mockHistoryService{}, nil)

	_, err := ctrl.List(context.Background(), &v1.DeviceHistoryListReq{DeviceNo: "  "})
	if err == nil {
		t.Fatal("expected error")
	}
	if gerror.Code(err) != gcode.CodeInvalidParameter {
		t.Fatalf("expected invalid parameter code, got %v", gerror.Code(err))
	}
}

func TestHistoryCtrlListReturnsServiceData(t *testing.T) {
	var gotDeviceNo string
	ctrl := NewHistoryCtrl(&mockHistoryService{
		listHistoryFn: func(ctx context.Context, deviceNo string) ([]entity.History, error) {
			gotDeviceNo = deviceNo
			return []entity.History{
				{Id: 101, DeviceNo: deviceNo, EventName: "喂奶"},
			}, nil
		},
	}, nil)

	res, err := ctrl.List(context.Background(), &v1.DeviceHistoryListReq{DeviceNo: "device-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotDeviceNo != "device-1" {
		t.Fatalf("unexpected deviceNo: %s", gotDeviceNo)
	}
	if len(res.List) != 1 || res.List[0].Id != 101 {
		t.Fatalf("unexpected response list: %+v", res.List)
	}
}

func TestHistoryCtrlSuggestDeleteValidationAndCall(t *testing.T) {
	called := false
	ctrl := NewHistoryCtrl(&mockHistoryService{
		deleteSuggestFn: func(ctx context.Context, id int64, deviceNo string) error {
			called = true
			if id != 9 || deviceNo != "dev-9" {
				t.Fatalf("unexpected args: id=%d deviceNo=%s", id, deviceNo)
			}
			return nil
		},
	}, nil)

	if _, err := ctrl.SuggestDelete(context.Background(), &v1.DeviceHistorySuggestDeleteReq{Id: 0, DeviceNo: "dev-9"}); err == nil {
		t.Fatal("expected invalid id error")
	}
	if _, err := ctrl.SuggestDelete(context.Background(), &v1.DeviceHistorySuggestDeleteReq{Id: 9, DeviceNo: ""}); err == nil {
		t.Fatal("expected missing deviceNo error")
	}

	if _, err := ctrl.SuggestDelete(context.Background(), &v1.DeviceHistorySuggestDeleteReq{Id: 9, DeviceNo: "dev-9"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected service delete call")
	}
}
