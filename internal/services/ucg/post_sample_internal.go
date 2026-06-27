package ucg

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"math"
	"strings"

	"hello/internal/dao"

	"github.com/gogf/gf/v2/database/gdb"
)

const (
	postSampleLimitDefault       = 20
	postSampleLimitMax           = 50
	postSampleRandomRecencyAlpha = 0.65 // 幂次偏置：略偏 high-id（≈新帖），1.0 为均匀
	postSampleExcludeMediaMax    = 8
	postSampleExcludeAuthorMax   = 10000 // 与 device sim/wx/ids 上限对齐，供 T6 排除 sim 作者
)

// PostSampleItem 内部抽样帖最小视图（无 author profile 富化）。
type PostSampleItem struct {
	PostId         uint64 `json:"postId"`
	AuthorWxId     int64  `json:"authorWxId"`
	Content        string `json:"content"`
	MediaType      int    `json:"mediaType"`
	CoverObjectKey string `json:"coverObjectKey,omitempty"`
	// CoverCdnUrl 供 sim 多模态 LLM 使用的封面 URL（图文全图 / 视频首帧 snapshot）。
	CoverCdnUrl string `json:"coverCdnUrl,omitempty"`
}

// SamplePublishedPosts 按 published_at 取最新已发布帖样本；单条有界 SQL，不调用 device/postsFromResult。
func SamplePublishedPosts(ctx context.Context, limit int, excludeMediaTypes []int, excludeAuthorWxIds []int64) ([]PostSampleItem, error) {
	if limit <= 0 {
		limit = postSampleLimitDefault
	}
	if limit > postSampleLimitMax {
		limit = postSampleLimitMax
	}
	rows, err := postSamplePublishedModel(ctx, excludeMediaTypes, excludeAuthorWxIds).
		OrderDesc("p." + dao.UcgPost.Columns().PublishedAt).
		OrderDesc("p." + dao.UcgPost.Columns().Id).
		Limit(limit).
		All()
	if err != nil {
		return nil, err
	}
	return postSampleFromRows(rows)
}

// SampleRandomPublishedPost 经 ID 探测从全库 published 帖随机取 1 条；锚点幂次偏置 α 略偏新帖。
func SampleRandomPublishedPost(ctx context.Context, excludeMediaTypes []int, excludeAuthorWxIds []int64) ([]PostSampleItem, error) {
	bounds, err := postSamplePublishedBoundsModel(ctx, excludeMediaTypes, excludeAuthorWxIds).
		Fields(
			"MIN("+dao.UcgPost.Columns().Id+") as min_id",
			"MAX("+dao.UcgPost.Columns().Id+") as max_id",
		).
		One()
	if err != nil {
		return nil, err
	}
	if bounds.IsEmpty() {
		return []PostSampleItem{}, nil
	}
	minID := bounds["min_id"].Uint64()
	maxID := bounds["max_id"].Uint64()
	if minID == 0 || maxID == 0 {
		return []PostSampleItem{}, nil
	}

	anchor := minID
	if minID < maxID {
		u, err := postSampleRandomUnit()
		if err != nil {
			return nil, err
		}
		span := float64(maxID - minID)
		anchor = minID + uint64(math.Floor(span*math.Pow(u, postSampleRandomRecencyAlpha)))
		if anchor > maxID {
			anchor = maxID
		}
	}

	rows, err := postSamplePublishedModel(ctx, excludeMediaTypes, excludeAuthorWxIds).
		Where("p."+dao.UcgPost.Columns().Id+" >= ?", anchor).
		OrderAsc("p." + dao.UcgPost.Columns().Id).
		Limit(1).
		All()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		// 锚点落在 id 空洞时回退 minId，保证 eligible 帖存在时必命中一条。
		rows, err = postSamplePublishedModel(ctx, excludeMediaTypes, excludeAuthorWxIds).
			Where("p."+dao.UcgPost.Columns().Id, minID).
			Limit(1).
			All()
		if err != nil {
			return nil, err
		}
	}
	return postSampleFromRows(rows)
}

func postSampleBaseModel(ctx context.Context) *gdb.Model {
	fields := postSampleSelectFields()
	args := make([]interface{}, len(fields))
	for i, f := range fields {
		args[i] = f
	}
	return dao.UcgPost.Ctx(ctx).As("p").
		Fields(args...).
		Where("p."+dao.UcgPost.Columns().Status, PostStatusPublished)
}

func postSamplePublishedModel(ctx context.Context, excludeMediaTypes []int, excludeAuthorWxIds []int64) *gdb.Model {
	m := postSampleApplyMediaExcludes(postSampleBaseModel(ctx), excludeMediaTypes)
	return postSampleApplyAuthorExcludes(m, excludeAuthorWxIds, "p."+dao.UcgPost.Columns().AuthorWxId)
}

func postSamplePublishedBoundsModel(ctx context.Context, excludeMediaTypes []int, excludeAuthorWxIds []int64) *gdb.Model {
	m := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Status, PostStatusPublished)
	m = postSampleApplyMediaExcludesOnTable(m, excludeMediaTypes, dao.UcgPost.Columns().MediaType)
	return postSampleApplyAuthorExcludes(m, excludeAuthorWxIds, dao.UcgPost.Columns().AuthorWxId)
}

func postSampleApplyMediaExcludes(m *gdb.Model, excludeMediaTypes []int) *gdb.Model {
	return postSampleApplyMediaExcludesOnTable(m, excludeMediaTypes, "p."+dao.UcgPost.Columns().MediaType)
}

func postSampleApplyMediaExcludesOnTable(m *gdb.Model, excludeMediaTypes []int, mediaTypeCol string) *gdb.Model {
	exclude := postSampleNormalizeExcludeMediaTypes(excludeMediaTypes)
	if len(exclude) == 0 {
		return m
	}
	args := make([]interface{}, len(exclude))
	for i, v := range exclude {
		args[i] = v
	}
	return m.WhereNotIn(mediaTypeCol, args)
}

func postSampleNormalizeExcludeMediaTypes(exclude []int) []int {
	if len(exclude) == 0 {
		return nil
	}
	seen := make(map[int]struct{}, len(exclude))
	out := make([]int, 0, len(exclude))
	for _, v := range exclude {
		if v < 0 {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
		if len(out) >= postSampleExcludeMediaMax {
			break
		}
	}
	return out
}

func postSampleApplyAuthorExcludes(m *gdb.Model, excludeAuthorWxIds []int64, authorCol string) *gdb.Model {
	exclude := postSampleNormalizeExcludeAuthorWxIds(excludeAuthorWxIds)
	if len(exclude) == 0 {
		return m
	}
	args := make([]interface{}, len(exclude))
	for i, id := range exclude {
		args[i] = id
	}
	return m.WhereNotIn(authorCol, args)
}

func postSampleNormalizeExcludeAuthorWxIds(exclude []int64) []int64 {
	if len(exclude) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(exclude))
	out := make([]int64, 0, len(exclude))
	for _, id := range exclude {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
		if len(out) >= postSampleExcludeAuthorMax {
			break
		}
	}
	return out
}

func postSampleSelectFields() []string {
	return []string{
		"p." + dao.UcgPost.Columns().Id + " as post_id",
		"p." + dao.UcgPost.Columns().AuthorWxId + " as author_wx_id",
		"p." + dao.UcgPost.Columns().Content + " as content",
		"p." + dao.UcgPost.Columns().MediaType + " as media_type",
		"(SELECT m." + dao.UcgPostMedia.Columns().ObjectKey +
			" FROM " + dao.UcgPostMedia.Table() + " m WHERE m." + dao.UcgPostMedia.Columns().PostId + "=p." + dao.UcgPost.Columns().Id +
			" ORDER BY m." + dao.UcgPostMedia.Columns().SortOrder + " ASC LIMIT 1) as cover_object_key",
	}
}

// postSampleRandomUnit 返回 [0,1) 均匀随机数，供幂次偏置锚点使用。
func postSampleRandomUnit() (float64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	// 53 位精度，避免 float64 低位精度损失。
	return float64(binary.BigEndian.Uint64(b[:])>>11) / float64(1<<53), nil
}

func postSampleFromRows(rows gdb.Result) ([]PostSampleItem, error) {
	out := make([]PostSampleItem, 0, len(rows))
	for _, row := range rows {
		item := PostSampleItem{
			PostId:     row["post_id"].Uint64(),
			AuthorWxId: row["author_wx_id"].Int64(),
			Content:    row["content"].String(),
			MediaType:  row["media_type"].Int(),
		}
		if key := strings.TrimSpace(row["cover_object_key"].String()); key != "" {
			item.CoverObjectKey = key
			item.CoverCdnUrl = postSampleCoverCdnURL(item.MediaType, key)
		}
		out = append(out, item)
	}
	return out, nil
}

// postSampleCoverCdnURL 按 mediaType 将封面 objectKey 转为 LLM 可访问的 CDN URL。
func postSampleCoverCdnURL(mediaType int, objectKey string) string {
	key := strings.TrimSpace(objectKey)
	if key == "" {
		return ""
	}
	switch mediaType {
	case MediaTypeImages:
		return BuildCdnURL(key)
	case MediaTypeVideo:
		return BuildVideoSnapshotURL(key)
	default:
		return ""
	}
}
