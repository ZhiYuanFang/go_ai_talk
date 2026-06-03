package ucg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// ProfileDTO App 侧 profile 视图。
type ProfileDTO struct {
	WxId         uint64 `json:"wxId"`
	Nickname     string `json:"nickname"`
	AvatarKey    string `json:"avatarKey"`
	AvatarUrl    string `json:"avatarUrl"`
	Bio          string `json:"bio"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
	AuditPending bool   `json:"auditPending,omitempty"`
	RejectReason string `json:"rejectReason,omitempty"`
}

// GetOrCreateMyProfile 获取当前用户 profile；不存在时经 device internal API 创建默认昵称。
func GetOrCreateMyProfile(ctx context.Context, wxID int64) (*ProfileDTO, error) {
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	exists, _, err := Device().ValidateWx(ctx, wxID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, gerror.NewCode(gcode.CodeNotFound, "wx 账号不存在")
	}
	row, err := dao.UcgProfile.Ctx(ctx).Where(dao.UcgProfile.Columns().WxId, wxID).One()
	if err != nil {
		return nil, err
	}
	if !row.IsEmpty() {
		var p entity.UcgProfile
		if err = row.Struct(&p); err != nil {
			return nil, err
		}
		return mergeProfileForAuthor(ctx, p)
	}
	babyName, err := Device().BabyName(ctx, wxID)
	if err != nil {
		return nil, err
	}
	nickname := defaultNickname(babyName)
	now := time.Now().Unix()
	res, err := dao.UcgProfile.Ctx(ctx).Data(g.Map{
		dao.UcgProfile.Columns().WxId:      wxID,
		dao.UcgProfile.Columns().Nickname:  nickname,
		dao.UcgProfile.Columns().CreatedAt: now,
		dao.UcgProfile.Columns().UpdatedAt: now,
	}).Insert()
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return profileToDTO(entity.UcgProfile{
		Id:        uint64(id),
		WxId:      uint64(wxID),
		Nickname:  nickname,
		CreatedAt: now,
		UpdatedAt: now,
	}), nil
}

// UpdateMyProfile 提交资料变更至 Green 待审队列；公开 profile 在通过前保持旧值。
func UpdateMyProfile(ctx context.Context, wxID int64, nickname, avatarKey, bio string) (*ProfileDTO, error) {
	base, err := GetOrCreateMyProfile(ctx, wxID)
	if err != nil {
		return nil, err
	}
	hasChange := strings.TrimSpace(nickname) != "" || strings.TrimSpace(avatarKey) != "" || strings.TrimSpace(bio) != ""
	if !hasChange {
		return base, nil
	}
	if err = EnqueueProfileAudit(ctx, wxID, nickname, avatarKey, bio); err != nil {
		return nil, err
	}
	row, err := dao.UcgProfile.Ctx(ctx).Where(dao.UcgProfile.Columns().WxId, wxID).One()
	if err != nil {
		return nil, err
	}
	var p entity.UcgProfile
	if err = row.Struct(&p); err != nil {
		return nil, err
	}
	return mergeProfileForAuthor(ctx, p)
}

// GetPublicProfile 公开 profile（不含敏感字段）。
func GetPublicProfile(ctx context.Context, wxID uint64) (*ProfileDTO, error) {
	row, err := dao.UcgProfile.Ctx(ctx).Where(dao.UcgProfile.Columns().WxId, wxID).One()
	if err != nil {
		return nil, err
	}
	if row.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotFound, "profile 不存在")
	}
	var p entity.UcgProfile
	if err = row.Struct(&p); err != nil {
		return nil, err
	}
	return profileToDTO(p), nil
}

func defaultNickname(babyName string) string {
	babyName = strings.TrimSpace(babyName)
	if babyName == "" {
		return "家长"
	}
	return fmt.Sprintf("%s的家长", babyName)
}

func profileToDTO(p entity.UcgProfile) *ProfileDTO {
	cfg := LoadOSSConfig(context.Background())
	avatarURL := ""
	if key := strings.TrimSpace(p.AvatarKey); key != "" {
		avatarURL = cfg.CdnBaseURL + "/" + strings.TrimPrefix(key, "/")
	}
	return &ProfileDTO{
		WxId:      p.WxId,
		Nickname:  p.Nickname,
		AvatarKey: p.AvatarKey,
		AvatarUrl: avatarURL,
		Bio:       p.Bio,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func mergeProfileForAuthor(ctx context.Context, p entity.UcgProfile) (*ProfileDTO, error) {
	dto := profileToDTO(p)
	reason, err := LoadProfileRejectReason(ctx, int64(p.WxId))
	if err != nil {
		return nil, err
	}
	dto.RejectReason = reason
	patch, ok, err := LoadProfilePending(ctx, int64(p.WxId))
	if err != nil {
		return nil, err
	}
	if !ok {
		return dto, nil
	}
	dto.AuditPending = true
	if patch.Nickname != "" {
		dto.Nickname = patch.Nickname
	}
	if patch.AvatarKey != "" {
		dto.AvatarKey = patch.AvatarKey
		cfg := LoadOSSConfig(ctx)
		dto.AvatarUrl = cfg.CdnBaseURL + "/" + strings.TrimPrefix(patch.AvatarKey, "/")
	}
	if patch.Bio != "" {
		dto.Bio = patch.Bio
	}
	if patch.UpdatedAt > dto.UpdatedAt {
		dto.UpdatedAt = patch.UpdatedAt
	}
	return dto, nil
}
