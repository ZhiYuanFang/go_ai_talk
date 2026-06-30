package ucg

import "context"

// BatchPublicProfilesForInternal 供 internal API 按 wxId 批量返回公开 profile 展示字段。
func BatchPublicProfilesForInternal(ctx context.Context, wxIDs []int64) ([]ProfileDTO, error) {
	uintIDs := make([]uint64, 0, len(wxIDs))
	for _, id := range wxIDs {
		if id > 0 {
			uintIDs = append(uintIDs, uint64(id))
		}
	}
	m, err := GetPublicProfilesByWxIDs(ctx, uintIDs)
	if err != nil {
		return nil, err
	}
	out := make([]ProfileDTO, 0, len(m))
	for _, dto := range m {
		if dto != nil {
			out = append(out, *dto)
		}
	}
	return out, nil
}
