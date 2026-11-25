package device

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const merchantFreeBabyQuota = 2

// MerchantCreateStaff 商户创建子账户：username/register + role=staff。
func MerchantCreateStaff(ctx context.Context, merchantID int64, account, password string) (int64, error) {
	if merchantID <= 0 {
		return 0, ErrMerchantNotFound
	}
	if _, err := MerchantByID(ctx, merchantID); err != nil {
		return 0, err
	}
	wxID, err := WxUsernameRegister(ctx, account, password)
	if err != nil {
		return 0, err
	}
	_, err = g.DB().Model("wx").Ctx(ctx).Where("id", wxID).Data(g.Map{
		"role": WxRoleStaff,
	}).Update()
	if err != nil {
		return 0, err
	}
	return wxID, nil
}

// MerchantCreateBaby 商户创建虚拟宝宝；无权益时最多 2 个。entitled 由调用方经 cash 探活传入。
func MerchantCreateBaby(ctx context.Context, merchantID int64, babyName string, birthdayUnixSec int64, sex int, entitled bool) (string, error) {
	if merchantID <= 0 {
		return "", ErrMerchantNotFound
	}
	if _, err := MerchantByID(ctx, merchantID); err != nil {
		return "", err
	}
	n, err := g.DB().Model("merchant_baby").Ctx(ctx).Where("merchant_id", merchantID).Count()
	if err != nil {
		return "", err
	}
	if !entitled && n >= merchantFreeBabyQuota {
		return "", ErrMerchantBabyQuota
	}
	dn, err := generateUniqueDeviceNo(ctx)
	if err != nil {
		return "", err
	}
	if _, err := DeviceAdmin().Register(ctx, dn); err != nil {
		return "", err
	}
	if err := DeviceAdmin().SaveUserProfile(ctx, dn, strings.TrimSpace(babyName), birthdayUnixSec, sex); err != nil {
		return "", err
	}
	now := time.Now().Unix()
	_, err = g.DB().Model("merchant_baby").Ctx(ctx).Data(g.Map{
		"merchant_id": merchantID,
		"device_no":   dn,
		"created_at":  now,
	}).Insert()
	if err != nil {
		return "", err
	}
	return dn, nil
}

// MerchantListBabies 商户名下宝宝 device_no 列表。
func MerchantListBabies(ctx context.Context, merchantID int64) ([]string, error) {
	rows, err := g.DB().Model("merchant_baby").Ctx(ctx).
		Where("merchant_id", merchantID).OrderAsc("id").All()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		dn := strings.TrimSpace(r["device_no"].String())
		if dn != "" {
			out = append(out, dn)
		}
	}
	return out, nil
}

// MerchantOwnsBaby 宝宝是否归属商户。
func MerchantOwnsBaby(ctx context.Context, merchantID int64, deviceNo string) (bool, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if merchantID <= 0 || deviceNo == "" {
		return false, nil
	}
	n, err := g.DB().Model("merchant_baby").Ctx(ctx).
		Where("merchant_id", merchantID).Where("device_no", deviceNo).Count()
	return n > 0, err
}

// MerchantOwnerOfDevice 返回拥有该 device 的 merchant_id；无则 0。
func MerchantOwnerOfDevice(ctx context.Context, deviceNo string) (int64, error) {
	deviceNo = strings.TrimSpace(deviceNo)
	if deviceNo == "" {
		return 0, nil
	}
	one, err := g.DB().Model("merchant_baby").Ctx(ctx).
		Fields("merchant_id").Where("device_no", deviceNo).Limit(1).One()
	if err != nil {
		return 0, err
	}
	if one.IsEmpty() {
		return 0, nil
	}
	return one["merchant_id"].Int64(), nil
}

// MerchantAssignBabyToStaff 分发宝宝给 staff；setFocus 时把该宝宝设为 staff 焦点。
func MerchantAssignBabyToStaff(ctx context.Context, merchantID, staffWxID int64, deviceNo string, setFocus bool) error {
	ok, err := MerchantOwnsBaby(ctx, merchantID, deviceNo)
	if err != nil {
		return err
	}
	if !ok {
		return ErrMerchantBabyNotOwned
	}
	role, err := wxRole(ctx, staffWxID)
	if err != nil {
		return err
	}
	if role != WxRoleStaff {
		return ErrAssignStaffRoleRequired
	}
	return AddBabyMember(ctx, staffWxID, strings.TrimSpace(deviceNo), BabyMemberSourceMerchantAssign, setFocus)
}

// MerchantUnassignBabyFromStaff 商户解绑 staff↔宝宝（仅删成员边）。
func MerchantUnassignBabyFromStaff(ctx context.Context, merchantID, staffWxID int64, deviceNo string) error {
	ok, err := MerchantOwnsBaby(ctx, merchantID, deviceNo)
	if err != nil {
		return err
	}
	if !ok {
		return ErrMerchantBabyNotOwned
	}
	deviceNo = strings.TrimSpace(deviceNo)
	res, err := g.DB().Model("wx_baby_member").Ctx(ctx).
		Where("wx_id", staffWxID).Where("device_no", deviceNo).Delete()
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrBabyMemberNotFound
	}
	// 若解绑的是焦点，尽量切到仍剩的第一个成员。
	focus, _ := WxDeviceNoByWxID(ctx, staffWxID)
	if focus == deviceNo {
		members, e := ListBabyMembers(ctx, staffWxID)
		if e == nil && len(members) > 0 {
			_ = SetFocusBaby(ctx, staffWxID, members[0].DeviceNo)
		} else {
			_, _ = g.DB().Model("wx").Ctx(ctx).Where("id", staffWxID).Data(g.Map{
				"device_no": "",
			}).Update()
		}
	}
	return nil
}

// MerchantUpdateBabyProfile 商户后台写画像。
func MerchantUpdateBabyProfile(ctx context.Context, merchantID int64, deviceNo, babyName string, birthdayUnixSec int64, sex int) error {
	ok, err := MerchantOwnsBaby(ctx, merchantID, deviceNo)
	if err != nil {
		return err
	}
	if !ok {
		return ErrMerchantBabyNotOwned
	}
	return DeviceAdmin().SaveUserProfile(ctx, strings.TrimSpace(deviceNo), strings.TrimSpace(babyName), birthdayUnixSec, sex)
}

// RefuseWxRoleChange 禁止 parent↔staff 变更。
func RefuseWxRoleChange(ctx context.Context, wxID int64, newRole string) error {
	_ = ctx
	_ = wxID
	_ = newRole
	return ErrWxRoleChangeForbidden
}
