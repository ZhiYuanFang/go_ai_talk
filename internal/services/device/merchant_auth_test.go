package device

import (
	"context"
	"errors"
	"testing"
)

func TestNormalizeStoreName(t *testing.T) {
	if _, err := normalizeStoreName("a"); !errors.Is(err, ErrMerchantStoreNameInvalid) {
		t.Fatalf("want invalid, got %v", err)
	}
	name, err := normalizeStoreName("  胖宝门店  ")
	if err != nil || name != "胖宝门店" {
		t.Fatalf("got %q %v", name, err)
	}
}

func TestAppUnbindRejected(t *testing.T) {
	err := AppUnbindBabyRejected(context.Background(), 1, "ABC123")
	if !errors.Is(err, ErrBabyUnbindNotAllowed) {
		t.Fatalf("want unbind rejected, got %v", err)
	}
}

func TestRefuseWxRoleChange(t *testing.T) {
	err := RefuseWxRoleChange(context.Background(), 1, WxRoleStaff)
	if !errors.Is(err, ErrWxRoleChangeForbidden) {
		t.Fatalf("want role change forbidden, got %v", err)
	}
}

func TestWxRoleConstants(t *testing.T) {
	if WxRoleParent != "parent" || WxRoleStaff != "staff" {
		t.Fatal("role constants")
	}
	if merchantFreeBabyQuota != 2 {
		t.Fatalf("free quota want 2 got %d", merchantFreeBabyQuota)
	}
}
