package ucg

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/platform/cachekit"

	"github.com/gogf/gf/v2/frame/g"
)

const profilePendingTTL = 7 * 24 * time.Hour

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
	key := cachekit.UCGProfilePendingDataKey(wxID)
	if err = ucgCache.SetEX(ctx, key, string(raw), profilePendingTTL); err != nil {
		return err
	}
	return ucgCache.SetAdd(ctx, cachekit.UCGProfilePendingSetKey(), strconv.FormatInt(wxID, 10))
}

// LoadProfilePending 读取作者待审资料 patch；无则 ok=false。
func LoadProfilePending(ctx context.Context, wxID int64) (patch ProfilePendingPatch, ok bool, err error) {
	key := cachekit.UCGProfilePendingDataKey(wxID)
	raw, hit, err := ucgCache.Get(ctx, key)
	if err != nil {
		return patch, false, err
	}
	if !hit || strings.TrimSpace(raw) == "" {
		return patch, false, nil
	}
	if err = json.Unmarshal([]byte(raw), &patch); err != nil {
		return patch, false, err
	}
	return patch, true, nil
}

// LoadProfileRejectReason 读取最近一次资料审核失败原因（仅作者可见）。
func LoadProfileRejectReason(ctx context.Context, wxID int64) (string, error) {
	key := cachekit.UCGProfileRejectReasonKey(wxID)
	raw, ok, err := ucgCache.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}
	return strings.TrimSpace(raw), nil
}

func listPendingProfileWxIDs(ctx context.Context, limit int) ([]int64, error) {
	if limit <= 0 {
		limit = 20
	}
	members, err := ucgCache.SetMembers(ctx, cachekit.UCGProfilePendingSetKey())
	if err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(members))
	for _, item := range members {
		if len(out) >= limit {
			break
		}
		id, pErr := strconv.ParseInt(strings.TrimSpace(item), 10, 64)
		if pErr != nil || id <= 0 {
			continue
		}
		out = append(out, id)
	}
	return out, nil
}

func clearProfilePending(ctx context.Context, wxID int64) error {
	id := strconv.FormatInt(wxID, 10)
	if err := ucgCache.Del(ctx, cachekit.UCGProfilePendingDataKey(wxID)); err != nil {
		return err
	}
	return ucgCache.SetRemove(ctx, cachekit.UCGProfilePendingSetKey(), id)
}

func setProfileRejectReason(ctx context.Context, wxID int64, reason string) error {
	key := cachekit.UCGProfileRejectReasonKey(wxID)
	return ucgCache.SetEX(ctx, key, strings.TrimSpace(reason), profilePendingTTL)
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
