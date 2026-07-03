package ucg

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// ForceTierFromValue 原力档位：[0,500) none；[500,1000) bronze；之后每 +500 一档至 diamond。
func ForceTierFromValue(v int) string {
	if v < 500 {
		return ""
	}
	if v < 1000 {
		return "bronze"
	}
	tiers := []string{"silver", "gold", "platinum", "diamond"}
	idx := (v - 1000) / 500
	if idx >= len(tiers) {
		return tiers[len(tiers)-1]
	}
	return tiers[idx]
}

func enrichProfileForceValues(ctx context.Context, dtos ...*ProfileDTO) {
	if len(dtos) == 0 {
		return
	}
	m := make(map[uint64]*ProfileDTO, len(dtos))
	for _, dto := range dtos {
		if dto != nil && dto.WxId > 0 {
			m[dto.WxId] = dto
		}
	}
	enrichProfileForceValuesMap(ctx, m)
}

func enrichProfileForceValuesMap(ctx context.Context, dtos map[uint64]*ProfileDTO) {
	if len(dtos) == 0 {
		return
	}
	wxIDs := make([]int64, 0, len(dtos))
	for id := range dtos {
		wxIDs = append(wxIDs, int64(id))
	}
	batch, err := Device().BatchWx(ctx, wxIDs)
	if err != nil {
		g.Log().Warningf(ctx, "[ucg-force] batch force_value failed err=%v", err)
		return
	}
	for id, dto := range dtos {
		if dto == nil {
			continue
		}
		if item, ok := batch[int64(id)]; ok && item.Exists {
			dto.ForceValue = item.ForceValue
			dto.ForceTier = ForceTierFromValue(item.ForceValue)
		}
	}
}

func enrichAuthorForceOnPosts(ctx context.Context, posts []*PostDTO) {
	if len(posts) == 0 {
		return
	}
	authors := make(map[uint64]*ProfileDTO)
	for _, p := range posts {
		if p != nil && p.Author != nil && p.Author.WxId > 0 {
			authors[p.Author.WxId] = p.Author
		}
	}
	enrichProfileForceValuesMap(ctx, authors)
}
