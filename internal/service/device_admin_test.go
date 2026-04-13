package service

import (
	"context"
	"errors"
	"testing"

	"hello/internal/model/entity"
)

func TestDeviceAdminVerifyPassword(t *testing.T) {
	if !DeviceAdmin().VerifyPassword("a521521521") {
		t.Fatal("expected password to be accepted")
	}
	if DeviceAdmin().VerifyPassword("wrong") {
		t.Fatal("expected password to be rejected")
	}
}

func TestDeviceAdminRegisterSuccessAndDuplicate(t *testing.T) {
	oldCount := userCountByDevice
	oldInsert := userInsertDevice
	defer func() {
		userCountByDevice = oldCount
		userInsertDevice = oldInsert
	}()

	existCount := 0
	userCountByDevice = func(ctx context.Context, deviceNo string) (int, error) {
		return existCount, nil
	}
	inserted := false
	userInsertDevice = func(ctx context.Context, deviceNo, activeTime string) error {
		inserted = true
		if deviceNo != "device-1" {
			t.Fatalf("unexpected deviceNo: %s", deviceNo)
		}
		if activeTime == "" {
			t.Fatal("activeTime should not be empty")
		}
		return nil
	}

	if _, err := DeviceAdmin().Register(context.Background(), "device-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !inserted {
		t.Fatal("expected insert to be called")
	}

	existCount = 1
	if _, err := DeviceAdmin().Register(context.Background(), "device-1"); !errors.Is(err, ErrDeviceExists) {
		t.Fatalf("expected ErrDeviceExists, got %v", err)
	}
}

func TestDeviceAdminListAndUpdateLastTalk(t *testing.T) {
	oldCount := userCountByDevice
	oldUpdate := userUpdateLastTalk
	oldList := userListAll
	defer func() {
		userCountByDevice = oldCount
		userUpdateLastTalk = oldUpdate
		userListAll = oldList
	}()

	userListAll = func(ctx context.Context) ([]entity.User, error) {
		return []entity.User{{
			DeviceNo:       "device-1",
			ActiveTime:     "2026-03-09 10:00:00",
			LastTalkTime:   "2026-03-09 10:10:00",
			LastTalkAsk:    "你好",
			LastTalkAnswer: "你好呀",
		}}, nil
	}

	items, err := DeviceAdmin().List(context.Background())
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("unexpected list length: %d", len(items))
	}
	if items[0].LastTalkAsk != "你好" || items[0].LastTalkAnswer != "你好呀" {
		t.Fatalf("unexpected ask/answer: %+v", items[0])
	}

	userCountByDevice = func(ctx context.Context, deviceNo string) (int, error) {
		return 1, nil
	}
	updated := false
	userUpdateLastTalk = func(ctx context.Context, deviceNo, lastTalkTime, ask, answer string) error {
		updated = true
		if deviceNo != "device-1" || ask != "问题" || answer != "回答" {
			t.Fatalf("unexpected update params: %s %s %s", deviceNo, ask, answer)
		}
		if lastTalkTime == "" {
			t.Fatal("lastTalkTime should not be empty")
		}
		return nil
	}
	if err := DeviceAdmin().UpdateLastTalk(context.Background(), "device-1", "问题", "回答"); err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}
	if !updated {
		t.Fatal("expected update to be called")
	}
}

func TestDeviceAdminUpdateLastTalkRejectsUnregistered(t *testing.T) {
	oldCount := userCountByDevice
	defer func() { userCountByDevice = oldCount }()

	userCountByDevice = func(ctx context.Context, deviceNo string) (int, error) {
		return 0, nil
	}
	if err := DeviceAdmin().UpdateLastTalk(context.Background(), "device-x", "a", "b"); !errors.Is(err, ErrDeviceNotRegistered) {
		t.Fatalf("expected ErrDeviceNotRegistered, got %v", err)
	}
}
