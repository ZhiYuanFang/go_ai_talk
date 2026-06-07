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
	WxId                uint64 `json:"wxId"`
	Nickname            string `json:"nickname"`
	AvatarKey           string `json:"avatarKey"`
	AvatarUrl           string `json:"avatarUrl"`
	AvatarThumbnailUrl  string `json:"avatarThumbnailUrl,omitempty"`
	Bio                 string `json:"bio"`
	FollowerCount  int    `json:"followerCount,omitempty"`
	FollowingCount int    `json:"followingCount,omitempty"`
	PostCount      int    `json:"postCount,omitempty"`
	CreatedAt      int64  `json:"createdAt"`
	UpdatedAt      int64  `json:"updatedAt"`
	AuditPending   bool   `json:"auditPending,omitempty"`
	RejectReason   string `json:"rejectReason,omitempty"`
	IpLocation     string `json:"ipLocation,omitempty"`
}

// GetOrCreateMyProfile 获取当前用户 profile；不存在时经 device internal API 创建默认昵称。
// clientIP 为网关注入的真实 IP，用于解析并节流更新 wx IP 属地。
func GetOrCreateMyProfile(ctx context.Context, wxID int64, clientIP string) (*ProfileDTO, error) {
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	exists, babyName, err := Device().ValidateWx(ctx, wxID)
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
		if err = refreshDefaultNicknameIfNeeded(ctx, wxID, &p); err != nil {
			g.Log().Warningf(ctx, "[ucg-profile] 刷新默认昵称失败 wxId=%d err=%v", wxID, err)
		}
		dto, err := mergeProfileForAuthor(ctx, p)
		if err != nil {
			return nil, err
		}
		if loc, locErr := MaybeUpdateWxIpLocation(ctx, wxID, clientIP); locErr == nil && loc != "" {
			dto.IpLocation = loc
		} else if loc == "" {
			if stored, _ := loadWxIpLocation(ctx, wxID); stored != "" {
				dto.IpLocation = stored
			}
		}
		return dto, nil
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
	dto := profileToDTO(entity.UcgProfile{
		Id:        uint64(id),
		WxId:      uint64(wxID),
		Nickname:  nickname,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if loc, locErr := MaybeUpdateWxIpLocation(ctx, wxID, clientIP); locErr == nil && loc != "" {
		dto.IpLocation = loc
	}
	return dto, nil
}

// UpdateMyProfile 提交资料变更至 Green 待审队列；公开 profile 在通过前保持旧值。
func UpdateMyProfile(ctx context.Context, wxID int64, nickname, avatarKey, bio string) (*ProfileDTO, error) {
	base, err := GetOrCreateMyProfile(ctx, wxID, "")
	if err != nil {
		return nil, err
	}
	hasChange := strings.TrimSpace(nickname) != "" || strings.TrimSpace(avatarKey) != "" || strings.TrimSpace(bio) != ""
	if !hasChange {
		return base, nil
	}
	if err = EnqueueProfileAudit(ctx, wxID, nickname, avatarKey, bio); err != nil {
		g.Log().Warningf(ctx, "[ucg-profile] 待审队列写入失败 wxId=%d err=%v，降级直写 profile", wxID, err)
		patch := ProfilePendingPatch{
			WxId:      wxID,
			Nickname:  strings.TrimSpace(nickname),
			AvatarKey: strings.TrimSpace(avatarKey),
			Bio:       strings.TrimSpace(bio),
			UpdatedAt: time.Now().Unix(),
		}
		if applyErr := applyProfilePending(ctx, patch); applyErr != nil {
			return nil, applyErr
		}
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
	if err = refreshDefaultNicknameIfNeeded(ctx, int64(wxID), &p); err != nil {
		g.Log().Warningf(ctx, "[ucg-profile] 刷新默认昵称失败 wxId=%d err=%v", wxID, err)
	}
	dto := profileToDTO(p)
	if locMap, locErr := IpLocationForWxIDs(ctx, []int64{int64(wxID)}); locErr == nil {
		if loc, ok := locMap[int64(wxID)]; ok {
			dto.IpLocation = loc
		}
	}
	return dto, nil
}

func isStoredDefaultNickname(nickname string) bool {
	nickname = strings.TrimSpace(nickname)
	return nickname == "" || nickname == "家长"
}

// refreshDefaultNicknameIfNeeded 当库内昵称为空或默认「家长」时，经 ValidateWx 取 babyName 并回写。
func refreshDefaultNicknameIfNeeded(ctx context.Context, wxID int64, p *entity.UcgProfile) error {
	if p == nil || wxID <= 0 || !isStoredDefaultNickname(p.Nickname) {
		return nil
	}
	_, babyName, err := Device().ValidateWx(ctx, wxID)
	if err != nil {
		return err
	}
	newNick := defaultNickname(babyName)
	if newNick == strings.TrimSpace(p.Nickname) {
		return nil
	}
	now := time.Now().Unix()
	_, err = dao.UcgProfile.Ctx(ctx).
		Where(dao.UcgProfile.Columns().WxId, wxID).
		Data(g.Map{
			dao.UcgProfile.Columns().Nickname:  newNick,
			dao.UcgProfile.Columns().UpdatedAt: now,
		}).Update()
	if err != nil {
		return err
	}
	p.Nickname = newNick
	p.UpdatedAt = now
	return nil
}

func defaultNickname(babyName string) string {
	babyName = strings.TrimSpace(babyName)
	if babyName == "" {
		return "家长"
	}
	return fmt.Sprintf("%s的家长", babyName)
}

func profileToDTO(p entity.UcgProfile) *ProfileDTO {
	key := strings.TrimSpace(p.AvatarKey)
	avatarURL := ""
	avatarThumbURL := ""
	if key != "" {
		avatarURL = BuildCdnURL(key)
		avatarThumbURL = BuildImageThumbnailURL(key)
	}
	return &ProfileDTO{
		WxId:               p.WxId,
		Nickname:           p.Nickname,
		AvatarKey:          p.AvatarKey,
		AvatarUrl:          avatarURL,
		AvatarThumbnailUrl: avatarThumbURL,
		Bio:                p.Bio,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}

func enrichProfileStats(ctx context.Context, wxID uint64, dto *ProfileDTO) {
	if dto == nil || wxID == 0 {
		return
	}
	if n, err := dao.UcgFollow.Ctx(ctx).
		Where(dao.UcgFollow.Columns().FollowerWxId, wxID).
		Count(); err == nil {
		dto.FollowingCount = int(n)
	}
	if n, err := dao.UcgFollow.Ctx(ctx).
		Where(dao.UcgFollow.Columns().FolloweeWxId, wxID).
		Count(); err == nil {
		dto.FollowerCount = int(n)
	}
	if n, err := dao.UcgPost.Ctx(ctx).
		Where(dao.UcgPost.Columns().AuthorWxId, wxID).
		Count(); err == nil {
		dto.PostCount = int(n)
	}
}

func mergeProfileForAuthor(ctx context.Context, p entity.UcgProfile) (*ProfileDTO, error) {
	dto := profileToDTO(p)
	enrichProfileStats(ctx, p.WxId, dto)
	reason, err := LoadProfileRejectReason(ctx, int64(p.WxId))
	if err != nil {
		g.Log().Warningf(ctx, "[ucg-profile] 读取资料审核拒绝原因失败 wxId=%d err=%v", p.WxId, err)
	} else {
		dto.RejectReason = reason
	}
	patch, ok, err := LoadProfilePending(ctx, int64(p.WxId))
	if err != nil {
		g.Log().Warningf(ctx, "[ucg-profile] 读取待审资料失败 wxId=%d err=%v", p.WxId, err)
		return dto, nil
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
		dto.AvatarUrl = BuildCdnURL(patch.AvatarKey)
		dto.AvatarThumbnailUrl = BuildImageThumbnailURL(patch.AvatarKey)
	}
	if patch.Bio != "" {
		dto.Bio = patch.Bio
	}
	if patch.UpdatedAt > dto.UpdatedAt {
		dto.UpdatedAt = patch.UpdatedAt
	}
	return dto, nil
}
