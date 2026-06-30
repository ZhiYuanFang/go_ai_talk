package simuser

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
)

// AdminUserListItem sim-admin 用户列表行。
type AdminUserListItem struct {
	WxId                int64  `json:"wxId"`
	Account             string `json:"account"`
	Nickname            string `json:"nickname"`
	AvatarUrl           string `json:"avatarUrl,omitempty"`
	AvatarThumbnailUrl  string `json:"avatarThumbnailUrl,omitempty"`
	CreatedAt           int64  `json:"createdAt"`
	PasswordPlain       string `json:"passwordPlain"`
	PasswordPlainLegacy bool   `json:"passwordPlainLegacy,omitempty"`
}

// AdminUserListResult 分页列表。
type AdminUserListResult struct {
	List     []AdminUserListItem `json:"list"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
}

type ucgProfileBatchItem struct {
	WxId               int64  `json:"wxId"`
	Nickname           string `json:"nickname"`
	AvatarUrl          string `json:"avatarUrl"`
	AvatarThumbnailUrl string `json:"avatarThumbnailUrl"`
}

// ListSimUsersForAdmin 编排 device list + ucg profiles + credential。
func ListSimUsersForAdmin(ctx context.Context, page, pageSize int) (AdminUserListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var data struct {
		List []struct {
			WxId    int64  `json:"wxId"`
			Account string `json:"account"`
		} `json:"list"`
		Total int `json:"total"`
	}
	path := fmt.Sprintf("/device/internal/api/sim/wx/list?page=%d&pageSize=%d", page, pageSize)
	if err := deviceInternalGet(ctx, path, &data); err != nil {
		return AdminUserListResult{}, err
	}
	wxIDs := make([]int64, 0, len(data.List))
	for _, row := range data.List {
		wxIDs = append(wxIDs, row.WxId)
	}
	profileByWx := map[int64]ucgProfileBatchItem{}
	if len(wxIDs) > 0 {
		var prof struct {
			List []ucgProfileBatchItem `json:"list"`
		}
		if err := ucgInternalPost(ctx, "/ucg/internal/api/profiles/batch", g.Map{"wxIds": wxIDs}, &prof); err != nil {
			return AdminUserListResult{}, err
		}
		for _, p := range prof.List {
			profileByWx[p.WxId] = p
		}
	}
	credMap, err := GetWxCredentialsByWxIDs(ctx, wxIDs)
	if err != nil {
		return AdminUserListResult{}, err
	}
	out := AdminUserListResult{
		Total: data.Total, Page: page, PageSize: pageSize,
		List: make([]AdminUserListItem, 0, len(data.List)),
	}
	for _, row := range data.List {
		item := AdminUserListItem{
			WxId: row.WxId, Account: row.Account,
		}
		if p, ok := profileByWx[row.WxId]; ok {
			item.Nickname = p.Nickname
			item.AvatarUrl = p.AvatarUrl
			item.AvatarThumbnailUrl = p.AvatarThumbnailUrl
		}
		if cred, ok := credMap[row.WxId]; ok {
			item.CreatedAt = cred.CreatedAt
			item.PasswordPlain = cred.PasswordPlain
		} else {
			item.PasswordPlain = legacySimDefaultPassword
			item.PasswordPlainLegacy = true
		}
		out.List = append(out.List, item)
	}
	return out, nil
}

// DeactivateSimUserForAdmin 注销单个模拟用户（仅删 wx + 本地清理）。
func DeactivateSimUserForAdmin(ctx context.Context, wxID int64) error {
	if wxID <= 0 {
		return fmt.Errorf("wxId 无效")
	}
	path := fmt.Sprintf("/device/internal/api/sim/wx/%d/deactivate", wxID)
	if err := deviceInternalPost(ctx, path, g.Map{}, nil); err != nil {
		return err
	}
	_ = SkipPendingVideoJobsForWx(ctx, wxID)
	return DeleteWxCredentialByWxID(ctx, wxID)
}
