package service

import (
	"context"
	"errors"
	"testing"

	"hello/internal/model/entity"
)

type fakeHistoryContract struct {
	listHistoryFn func(ctx context.Context, deviceNo string) ([]entity.History, error)
}

func (f *fakeHistoryContract) ListHistory(ctx context.Context, deviceNo string) ([]entity.History, error) {
	if f.listHistoryFn != nil {
		return f.listHistoryFn(ctx, deviceNo)
	}
	return nil, nil
}
func (f *fakeHistoryContract) ListSuggest(ctx context.Context, deviceNo string) ([]entity.Suggest, error) {
	return nil, nil
}
func (f *fakeHistoryContract) DeleteSuggest(ctx context.Context, id int64, deviceNo string) error {
	return nil
}
func (f *fakeHistoryContract) ListEventOptions(ctx context.Context) ([]entity.Event, error) {
	return nil, nil
}
func (f *fakeHistoryContract) GetBirthday(ctx context.Context, deviceNo string) (string, int, error) {
	return "", 0, nil
}
func (f *fakeHistoryContract) SaveBirthday(ctx context.Context, deviceNo, birthday string, sex int) error {
	return nil
}
func (f *fakeHistoryContract) AddHistory(ctx context.Context, item entity.History) (int64, error) {
	return 0, nil
}
func (f *fakeHistoryContract) UpdateHistory(ctx context.Context, item entity.History) error {
	return nil
}
func (f *fakeHistoryContract) DeleteHistory(ctx context.Context, id int64, deviceNo string) error {
	return nil
}

func TestHistorySwitchAdapterLocalMode(t *testing.T) {
	localCalled := 0
	remoteCalled := 0
	local := &fakeHistoryContract{
		listHistoryFn: func(ctx context.Context, deviceNo string) ([]entity.History, error) {
			localCalled++
			return []entity.History{{Id: 1}}, nil
		},
	}
	remote := &fakeHistoryContract{
		listHistoryFn: func(ctx context.Context, deviceNo string) ([]entity.History, error) {
			remoteCalled++
			return []entity.History{{Id: 2}}, nil
		},
	}
	adapter := newHistorySwitchAdapter(local, remote, historySwitchConfig{mode: historyModeLocal, failoverToLocal: true})

	res, err := adapter.ListHistory(context.Background(), "dev-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 || res[0].Id != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}
	if localCalled != 1 || remoteCalled != 0 {
		t.Fatalf("unexpected call count local=%d remote=%d", localCalled, remoteCalled)
	}
}

func TestHistorySwitchAdapterRemoteFailover(t *testing.T) {
	localCalled := 0
	remoteCalled := 0
	local := &fakeHistoryContract{
		listHistoryFn: func(ctx context.Context, deviceNo string) ([]entity.History, error) {
			localCalled++
			return []entity.History{{Id: 10}}, nil
		},
	}
	remote := &fakeHistoryContract{
		listHistoryFn: func(ctx context.Context, deviceNo string) ([]entity.History, error) {
			remoteCalled++
			return nil, errors.New("remote unavailable")
		},
	}
	adapter := newHistorySwitchAdapter(local, remote, historySwitchConfig{mode: historyModeRemote, failoverToLocal: true})

	res, err := adapter.ListHistory(context.Background(), "dev-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 || res[0].Id != 10 {
		t.Fatalf("unexpected result: %+v", res)
	}
	if localCalled != 1 || remoteCalled != 1 {
		t.Fatalf("unexpected call count local=%d remote=%d", localCalled, remoteCalled)
	}
}

func TestHistorySwitchAdapterRemoteNoFailover(t *testing.T) {
	local := &fakeHistoryContract{}
	remote := &fakeHistoryContract{
		listHistoryFn: func(ctx context.Context, deviceNo string) ([]entity.History, error) {
			return nil, errors.New("remote unavailable")
		},
	}
	adapter := newHistorySwitchAdapter(local, remote, historySwitchConfig{mode: historyModeRemote, failoverToLocal: false})

	_, err := adapter.ListHistory(context.Background(), "dev-1")
	if err == nil {
		t.Fatal("expected remote error")
	}
}

func TestHistorySwitchAdapterCanaryPercent(t *testing.T) {
	localCalled := 0
	remoteCalled := 0
	local := &fakeHistoryContract{
		listHistoryFn: func(ctx context.Context, deviceNo string) ([]entity.History, error) {
			localCalled++
			return []entity.History{{Id: 1}}, nil
		},
	}
	remote := &fakeHistoryContract{
		listHistoryFn: func(ctx context.Context, deviceNo string) ([]entity.History, error) {
			remoteCalled++
			return []entity.History{{Id: 2}}, nil
		},
	}

	adapterCanary0 := newHistorySwitchAdapter(local, remote, historySwitchConfig{mode: historyModeCanary, canaryPercent: 0, failoverToLocal: true})
	_, _ = adapterCanary0.ListHistory(context.Background(), "dev-a")

	adapterCanary100 := newHistorySwitchAdapter(local, remote, historySwitchConfig{mode: historyModeCanary, canaryPercent: 100, failoverToLocal: true})
	_, _ = adapterCanary100.ListHistory(context.Background(), "dev-b")

	if localCalled != 1 || remoteCalled != 1 {
		t.Fatalf("unexpected call count local=%d remote=%d", localCalled, remoteCalled)
	}
}
