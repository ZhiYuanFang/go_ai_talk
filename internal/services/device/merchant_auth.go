package device

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrMerchantStoreNameInvalid = errors.New("店名格式无效，请输入 2-32 个字符")
	ErrMerchantPasswordInvalid  = errors.New("密码格式无效，请输入 6-64 位字符")
	ErrMerchantStoreNameTaken   = errors.New("店名已存在")
	ErrMerchantAuthFailed       = errors.New("店名或密码错误")
	ErrMerchantNotFound         = errors.New("商户不存在")
)

// MerchantRow 商户主账号行。
type MerchantRow struct {
	ID           int64  `json:"id"`
	StoreName    string `json:"storeName"`
	PasswordHash string `json:"-"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
}

func normalizeStoreName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	n := utf8.RuneCountInString(name)
	if n < 2 || n > 32 {
		return "", ErrMerchantStoreNameInvalid
	}
	return name, nil
}

// MerchantRegister 店名+密码注册商户。
func MerchantRegister(ctx context.Context, storeName, password string) (int64, error) {
	name, err := normalizeStoreName(storeName)
	if err != nil {
		return 0, err
	}
	password, err = validatePassword(password)
	if err != nil {
		return 0, ErrMerchantPasswordInvalid
	}
	exist, err := g.DB().Model("merchant").Ctx(ctx).Where("store_name", name).Limit(1).One()
	if err != nil {
		return 0, err
	}
	if !exist.IsEmpty() {
		return 0, ErrMerchantStoreNameTaken
	}
	hash, err := hashPassword(password)
	if err != nil {
		return 0, err
	}
	now := time.Now().Unix()
	res, err := g.DB().Model("merchant").Ctx(ctx).Data(g.Map{
		"store_name":    name,
		"password_hash": hash,
		"created_at":    now,
		"updated_at":    now,
	}).Insert()
	if err != nil {
		if again, e2 := g.DB().Model("merchant").Ctx(ctx).Where("store_name", name).Limit(1).One(); e2 == nil && !again.IsEmpty() {
			return 0, ErrMerchantStoreNameTaken
		}
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// MerchantLogin 店名密码登录，返回商户 id。
func MerchantLogin(ctx context.Context, storeName, password string) (*MerchantRow, error) {
	name, err := normalizeStoreName(storeName)
	if err != nil {
		return nil, ErrMerchantAuthFailed
	}
	password = strings.TrimSpace(password)
	if password == "" {
		return nil, ErrMerchantAuthFailed
	}
	one, err := g.DB().Model("merchant").Ctx(ctx).Where("store_name", name).Limit(1).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return nil, ErrMerchantAuthFailed
	}
	hash := one["password_hash"].String()
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, ErrMerchantAuthFailed
	}
	return &MerchantRow{
		ID:        one["id"].Int64(),
		StoreName: one["store_name"].String(),
		CreatedAt: one["created_at"].Int64(),
		UpdatedAt: one["updated_at"].Int64(),
	}, nil
}

// MerchantByID 按主键读商户。
func MerchantByID(ctx context.Context, merchantID int64) (*MerchantRow, error) {
	if merchantID <= 0 {
		return nil, ErrMerchantNotFound
	}
	one, err := g.DB().Model("merchant").Ctx(ctx).Where("id", merchantID).Limit(1).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return nil, ErrMerchantNotFound
	}
	return &MerchantRow{
		ID:        one["id"].Int64(),
		StoreName: one["store_name"].String(),
		CreatedAt: one["created_at"].Int64(),
		UpdatedAt: one["updated_at"].Int64(),
	}, nil
}
