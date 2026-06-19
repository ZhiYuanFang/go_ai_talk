package ucg

import (
	"context"
	"strings"

	"hello/internal/dao"

	"github.com/gogf/gf/v2/database/gdb"
)

const (
	postSampleLimitDefault = 20
	postSampleLimitMax     = 50
)

// PostSampleItem 内部抽样帖最小视图（无 author/media 富化）。
type PostSampleItem struct {
	PostId         uint64 `json:"postId"`
	Content        string `json:"content"`
	MediaType      int    `json:"mediaType"`
	CoverObjectKey string `json:"coverObjectKey,omitempty"`
}

// SamplePublishedPosts 按 published_at 取最新已发布帖样本；单条有界 SQL，不调用 device/postsFromResult。
func SamplePublishedPosts(ctx context.Context, limit int) ([]PostSampleItem, error) {
	if limit <= 0 {
		limit = postSampleLimitDefault
	}
	if limit > postSampleLimitMax {
		limit = postSampleLimitMax
	}
	rows, err := dao.UcgPost.Ctx(ctx).As("p").
		Fields(
			"p."+dao.UcgPost.Columns().Id+" as post_id",
			"p."+dao.UcgPost.Columns().Content+" as content",
			"p."+dao.UcgPost.Columns().MediaType+" as media_type",
			"(SELECT m."+dao.UcgPostMedia.Columns().ObjectKey+
				" FROM "+dao.UcgPostMedia.Table()+" m WHERE m."+dao.UcgPostMedia.Columns().PostId+"=p."+dao.UcgPost.Columns().Id+
				" ORDER BY m."+dao.UcgPostMedia.Columns().SortOrder+" ASC LIMIT 1) as cover_object_key",
		).
		Where("p."+dao.UcgPost.Columns().Status, PostStatusPublished).
		OrderDesc("p." + dao.UcgPost.Columns().PublishedAt).
		OrderDesc("p." + dao.UcgPost.Columns().Id).
		Limit(limit).
		All()
	if err != nil {
		return nil, err
	}
	return postSampleFromRows(rows)
}

func postSampleFromRows(rows gdb.Result) ([]PostSampleItem, error) {
	out := make([]PostSampleItem, 0, len(rows))
	for _, row := range rows {
		item := PostSampleItem{
			PostId:    row["post_id"].Uint64(),
			Content:   row["content"].String(),
			MediaType: row["media_type"].Int(),
		}
		if key := strings.TrimSpace(row["cover_object_key"].String()); key != "" {
			item.CoverObjectKey = key
		}
		out = append(out, item)
	}
	return out, nil
}
