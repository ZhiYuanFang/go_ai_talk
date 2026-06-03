package ucg

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hello/internal/dao"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	redisProfilePendingSetKey = "ucg:green:profile:pending"
	redisProfilePendingPrefix = "ucg:green:profile:data:"
	redisProfileRejectPrefix  = "ucg:green:profile:reject:"
	profilePendingTTL         = 7 * 24 * time.Hour
)

// ProfilePendingPatch 待 Green 审核的资料变更（公开 API 不可见直至通过）。
type ProfilePendingPatch struct {
	WxId      int64  `json:"wxId"`
	Nickname  string `json:"nickname,omitempty"`
	AvatarKey string `json:"avatarKey,omitempty"`
	Bio       string `json:"bio,omitempty"`
	UpdatedAt int64  `json:"updatedAt"`
}

// EnqueueProfileAudit 将资料变更写入 Redis 待审队列，不立即对公众可见。
func EnqueueProfileAudit(ctx context.Context, wxID int64, nickname, avatarKey, bio string) error {
	if wxID <= 0 {
		return fmt.Errorf("wxId 无效")
	}
	patch := ProfilePendingPatch{
		WxId:      wxID,
		Nickname:  strings.TrimSpace(nickname),
		AvatarKey: strings.TrimSpace(avatarKey),
		Bio:       strings.TrimSpace(bio),
		UpdatedAt: time.Now().Unix(),
	}
	raw, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	key := redisProfilePendingPrefix + strconv.FormatInt(wxID, 10)
	if _, err = g.Redis().Do(ctx, "SET", key, string(raw), "EX", int(profilePendingTTL.Seconds())); err != nil {
		return err
	}
	_, err = g.Redis().Do(ctx, "SADD", redisProfilePendingSetKey, strconv.FormatInt(wxID, 10))
	return err
}

// LoadProfilePending 读取作者待审资料 patch；无则 ok=false。
func LoadProfilePending(ctx context.Context, wxID int64) (patch ProfilePendingPatch, ok bool, err error) {
	key := redisProfilePendingPrefix + strconv.FormatInt(wxID, 10)
	raw, err := g.Redis().Do(ctx, "GET", key)
	if err != nil {
		return patch, false, err
	}
	s := strings.TrimSpace(raw.String())
	if s == "" {
		return patch, false, nil
	}
	if err = json.Unmarshal([]byte(s), &patch); err != nil {
		return patch, false, err
	}
	return patch, true, nil
}

// LoadProfileRejectReason 读取最近一次资料审核失败原因（仅作者可见）。
func LoadProfileRejectReason(ctx context.Context, wxID int64) (string, error) {
	key := redisProfileRejectPrefix + strconv.FormatInt(wxID, 10)
	raw, err := g.Redis().Do(ctx, "GET", key)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(raw.String()), nil
}

func listPendingProfileWxIDs(ctx context.Context, limit int) ([]int64, error) {
	if limit <= 0 {
		limit = 20
	}
	raw, err := g.Redis().Do(ctx, "SMEMBERS", redisProfilePendingSetKey)
	if err != nil {
		return nil, err
	}
	arr := raw.Array()
	out := make([]int64, 0, len(arr))
	for _, item := range arr {
		if len(out) >= limit {
			break
		}
		id, pErr := strconv.ParseInt(strings.TrimSpace(g.NewVar(item).String()), 10, 64)
		if pErr != nil || id <= 0 {
			continue
		}
		out = append(out, id)
	}
	return out, nil
}

func clearProfilePending(ctx context.Context, wxID int64) error {
	id := strconv.FormatInt(wxID, 10)
	if _, err := g.Redis().Do(ctx, "DEL", redisProfilePendingPrefix+id); err != nil {
		return err
	}
	_, err := g.Redis().Do(ctx, "SREM", redisProfilePendingSetKey, id)
	return err
}

func setProfileRejectReason(ctx context.Context, wxID int64, reason string) error {
	key := redisProfileRejectPrefix + strconv.FormatInt(wxID, 10)
	_, err := g.Redis().Do(ctx, "SET", key, strings.TrimSpace(reason), "EX", int(profilePendingTTL.Seconds()))
	return err
}

func applyProfilePending(ctx context.Context, patch ProfilePendingPatch) error {
	data := g.Map{dao.UcgProfile.Columns().UpdatedAt: time.Now().Unix()}
	if patch.Nickname != "" {
		data[dao.UcgProfile.Columns().Nickname] = patch.Nickname
	}
	if patch.AvatarKey != "" {
		data[dao.UcgProfile.Columns().AvatarKey] = patch.AvatarKey
	}
	if patch.Bio != "" {
		data[dao.UcgProfile.Columns().Bio] = patch.Bio
	}
	_, err := dao.UcgProfile.Ctx(ctx).Where(dao.UcgProfile.Columns().WxId, patch.WxId).Data(data).Update()
	return err
}
