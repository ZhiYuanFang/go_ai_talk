package ucg

import (
	deviceclient "hello/internal/clients/device"
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// displayNicknameForWxWithoutProfile 无 ucg_profile 行时推导展示昵称（不写库）。
func displayNicknameForWxWithoutProfile(ctx context.Context, wxID int64) string {
	_, babyName, err := deviceclient.UcgAPI().ValidateWx(ctx, wxID)
	if err != nil {
		g.Log().Warningf(ctx, "[ucg-profile] 推导昵称失败 wxId=%d err=%v", wxID, err)
		return defaultNickname("")
	}
	return defaultNickname(babyName)
}

// BatchPublicProfilesForInternal 供 internal API 按 wxId 批量返回公开 profile 展示字段。
// 请求中每个有效 wxId 均在 list 返回一条；无 profile 行时返回推导昵称。
func BatchPublicProfilesForInternal(ctx context.Context, wxIDs []int64) ([]ProfileDTO, error) {
	unique := make([]int64, 0, len(wxIDs))
	seen := make(map[int64]struct{}, len(wxIDs))
	for _, id := range wxIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return []ProfileDTO{}, nil
	}
	uintIDs := make([]uint64, 0, len(unique))
	for _, id := range unique {
		uintIDs = append(uintIDs, uint64(id))
	}
	m, err := GetPublicProfilesByWxIDs(ctx, uintIDs)
	if err != nil {
		return nil, err
	}
	out := make([]ProfileDTO, 0, len(unique))
	for _, id := range unique {
		if dto, ok := m[uint64(id)]; ok && dto != nil {
			out = append(out, *dto)
			continue
		}
		out = append(out, ProfileDTO{
			WxId:     uint64(id),
			Nickname: displayNicknameForWxWithoutProfile(ctx, id),
		})
	}
	return out, nil
}
