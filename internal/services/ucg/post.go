package ucg

import (
	"context"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// PostMediaInput 创建/更新帖子时的媒体引用（仅存 objectKey）。
type PostMediaInput struct {
	ObjectKey  string `json:"objectKey"`
	MediaKind  int    `json:"mediaKind"`
	SortOrder  int    `json:"sortOrder"`
	DurationMs int    `json:"durationMs"`
	SizeBytes  int64  `json:"sizeBytes"`
}

// PostDTO 帖子视图。
type PostDTO struct {
	Id           uint64         `json:"id"`
	AuthorWxId   uint64         `json:"authorWxId"`
	Content      string         `json:"content"`
	Status       int            `json:"status"`
	RejectReason string         `json:"rejectReason,omitempty"`
	MediaType    int            `json:"mediaType"`
	LikeCount    uint           `json:"likeCount"`
	CommentCount uint           `json:"commentCount"`
	LikedByMe    bool           `json:"likedByMe"`
	CreatedAt    int64          `json:"createdAt"`
	UpdatedAt    int64          `json:"updatedAt"`
	PublishedAt  int64          `json:"publishedAt,omitempty"`
	IpLocation   string         `json:"ipLocation,omitempty"`
	Media        []PostMediaDTO `json:"media,omitempty"`
	Author       *ProfileDTO    `json:"author,omitempty"`
}

// PostMediaDTO 帖子媒体展示。
type PostMediaDTO struct {
	ObjectKey    string `json:"objectKey"`
	CdnUrl       string `json:"cdnUrl"`
	ThumbnailUrl string `json:"thumbnailUrl,omitempty"`
	MediaKind    int    `json:"mediaKind"`
	SortOrder    int    `json:"sortOrder"`
}

// CreatePost 创建帖子；submit=true 时进入 pending_audit。clientIP 用于服务端快照 ip_location。
func CreatePost(ctx context.Context, wxID int64, content string, mediaType int, submit bool, media []PostMediaInput, clientIP string) (*PostDTO, error) {
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	status := PostStatusDraft
	if submit {
		status = PostStatusPendingAudit
	}
	now := time.Now().Unix()
	ipLoc := SnapshotPostIpLocation(ctx, clientIP)
	data := g.Map{
		dao.UcgPost.Columns().AuthorWxId: wxID,
		dao.UcgPost.Columns().Content:    strings.TrimSpace(content),
		dao.UcgPost.Columns().Status:     status,
		dao.UcgPost.Columns().MediaType:  mediaType,
		dao.UcgPost.Columns().CreatedAt:  now,
		dao.UcgPost.Columns().UpdatedAt:  now,
	}
	if ipLoc != "" {
		data[dao.UcgPost.Columns().IpLocation] = ipLoc
	}
	res, err := dao.UcgPost.Ctx(ctx).Data(data).Insert()
	if err != nil {
		return nil, err
	}
	postID, _ := res.LastInsertId()
	if err = replacePostMedia(ctx, uint64(postID), media); err != nil {
		return nil, err
	}
	return GetPostByID(ctx, uint64(postID), wxID)
}

// UpdatePost 编辑自己的帖子。
func UpdatePost(ctx context.Context, wxID int64, postID uint64, content string, mediaType int, submit bool, media []PostMediaInput) (*PostDTO, error) {
	post, err := loadOwnedPost(ctx, wxID, postID)
	if err != nil {
		return nil, err
	}
	status := post.Status
	if submit || status == PostStatusPublished {
		status = PostStatusPendingAudit
	}
	now := time.Now().Unix()
	_, err = dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).Data(g.Map{
		dao.UcgPost.Columns().Content:   strings.TrimSpace(content),
		dao.UcgPost.Columns().Status:    status,
		dao.UcgPost.Columns().MediaType: mediaType,
		dao.UcgPost.Columns().UpdatedAt: now,
	}).Update()
	if err != nil {
		return nil, err
	}
	if err = replacePostMedia(ctx, postID, media); err != nil {
		return nil, err
	}
	return GetPostByID(ctx, postID, wxID)
}

// DeletePost 删除自己的帖子。
func DeletePost(ctx context.Context, wxID int64, postID uint64) error {
	if _, err := loadOwnedPost(ctx, wxID, postID); err != nil {
		return err
	}
	if _, err := dao.UcgPostMedia.Ctx(ctx).Where(dao.UcgPostMedia.Columns().PostId, postID).Delete(); err != nil {
		return err
	}
	_, err := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).Delete()
	return err
}

// ListMyPosts 我的动态（含全 status）。
func ListMyPosts(ctx context.Context, wxID int64, page, pageSize int) (*PageResult, error) {
	p := NormalizePage(page, pageSize)
	model := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().AuthorWxId, wxID)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows, err := model.OrderDesc(dao.UcgPost.Columns().CreatedAt).Limit(p.PageSize).Offset(pageOffset(p)).All()
	if err != nil {
		return nil, err
	}
	list, err := postsFromResult(ctx, rows, wxID)
	if err != nil {
		return nil, err
	}
	return &PageResult{List: list, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// ListUserPosts 指定作者已发布动态（status=published）。
func ListUserPosts(ctx context.Context, authorWxID, viewerWxID int64, page, pageSize int) (*PageResult, error) {
	if authorWxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	p := NormalizePage(page, pageSize)
	model := dao.UcgPost.Ctx(ctx).
		Where(dao.UcgPost.Columns().AuthorWxId, authorWxID).
		Where(dao.UcgPost.Columns().Status, PostStatusPublished)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows, err := model.OrderDesc(dao.UcgPost.Columns().CreatedAt).Limit(p.PageSize).Offset(pageOffset(p)).All()
	if err != nil {
		return nil, err
	}
	list, err := postsFromResult(ctx, rows, viewerWxID)
	if err != nil {
		return nil, err
	}
	return &PageResult{List: list, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

// GetPostByID 获取单帖；非 published 仅作者可见；enrich likedByMe。
func GetPostByID(ctx context.Context, postID uint64, viewerWxID int64) (*PostDTO, error) {
	row, err := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).One()
	if err != nil {
		return nil, err
	}
	if row.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotFound, "帖子不存在")
	}
	var post entity.UcgPost
	if err = row.Struct(&post); err != nil {
		return nil, err
	}
	if post.Status != PostStatusPublished && int64(post.AuthorWxId) != viewerWxID {
		return nil, gerror.NewCode(gcode.CodeNotFound, "帖子不存在")
	}
	dto, err := postToDTO(ctx, post)
	if err != nil {
		return nil, err
	}
	if viewerWxID > 0 {
		liked, lErr := likedPostIDSet(ctx, viewerWxID, []uint64{postID})
		if lErr == nil {
			_, dto.LikedByMe = liked[postID]
		}
	}
	return dto, nil
}

func loadOwnedPost(ctx context.Context, wxID int64, postID uint64) (*entity.UcgPost, error) {
	row, err := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).One()
	if err != nil {
		return nil, err
	}
	if row.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotFound, "帖子不存在")
	}
	var post entity.UcgPost
	if err = row.Struct(&post); err != nil {
		return nil, err
	}
	if int64(post.AuthorWxId) != wxID {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "无权操作该帖子")
	}
	return &post, nil
}

func replacePostMedia(ctx context.Context, postID uint64, media []PostMediaInput) error {
	if _, err := dao.UcgPostMedia.Ctx(ctx).Where(dao.UcgPostMedia.Columns().PostId, postID).Delete(); err != nil {
		return err
	}
	for i, m := range media {
		key := strings.TrimSpace(m.ObjectKey)
		if key == "" {
			continue
		}
		sortOrder := m.SortOrder
		if sortOrder == 0 {
			sortOrder = i
		}
		if _, err := dao.UcgPostMedia.Ctx(ctx).Data(g.Map{
			dao.UcgPostMedia.Columns().PostId:     postID,
			dao.UcgPostMedia.Columns().ObjectKey:  key,
			dao.UcgPostMedia.Columns().MediaKind:  m.MediaKind,
			dao.UcgPostMedia.Columns().SortOrder:  sortOrder,
			dao.UcgPostMedia.Columns().DurationMs: m.DurationMs,
			dao.UcgPostMedia.Columns().SizeBytes:  m.SizeBytes,
		}).Insert(); err != nil {
			return err
		}
	}
	return nil
}

func postsFromResult(ctx context.Context, rows gdb.Result, viewerWxID int64) ([]*PostDTO, error) {
	out := make([]*PostDTO, 0, len(rows))
	postIDs := make([]uint64, 0, len(rows))
	for _, row := range rows {
		var post entity.UcgPost
		if err := row.Struct(&post); err != nil {
			return nil, err
		}
		dto, err := postToDTO(ctx, post)
		if err != nil {
			return nil, err
		}
		out = append(out, dto)
		postIDs = append(postIDs, post.Id)
	}
	if viewerWxID > 0 && len(postIDs) > 0 {
		liked, err := likedPostIDSet(ctx, viewerWxID, postIDs)
		if err != nil {
			return nil, err
		}
		for _, dto := range out {
			_, dto.LikedByMe = liked[dto.Id]
		}
	}
	return out, nil
}

func likedPostIDSet(ctx context.Context, viewerWxID int64, postIDs []uint64) (map[uint64]struct{}, error) {
	out := make(map[uint64]struct{}, len(postIDs))
	if viewerWxID <= 0 || len(postIDs) == 0 {
		return out, nil
	}
	rows, err := dao.UcgPostLike.Ctx(ctx).
		Where(dao.UcgPostLike.Columns().WxId, viewerWxID).
		WhereIn(dao.UcgPostLike.Columns().PostId, postIDs).
		Fields(dao.UcgPostLike.Columns().PostId).
		All()
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		var like entity.UcgPostLike
		if err = row.Struct(&like); err != nil {
			return nil, err
		}
		out[like.PostId] = struct{}{}
	}
	return out, nil
}

func postToDTO(ctx context.Context, post entity.UcgPost) (*PostDTO, error) {
	media, err := loadPostMedia(ctx, post.Id)
	if err != nil {
		return nil, err
	}
	var author *ProfileDTO
	if prof, pErr := GetPublicProfile(ctx, post.AuthorWxId); pErr == nil {
		author = prof
	}
	ensureAuthorBio(author)
	dto := &PostDTO{
		Id:           post.Id,
		AuthorWxId:   post.AuthorWxId,
		Content:      post.Content,
		Status:       post.Status,
		RejectReason: post.RejectReason,
		MediaType:    post.MediaType,
		LikeCount:    post.LikeCount,
		CommentCount: post.CommentCount,
		CreatedAt:    post.CreatedAt,
		UpdatedAt:    post.UpdatedAt,
		PublishedAt:  post.PublishedAt,
		IpLocation:   strings.TrimSpace(post.IpLocation),
		Media:        media,
		Author:       author,
	}
	return dto, nil
}

// ensureAuthorBio 保证 author.bio 非空（Feed/详情展示）。
func ensureAuthorBio(author *ProfileDTO) {
	if author == nil {
		return
	}
	if strings.TrimSpace(author.Bio) == "" {
		author.Bio = " "
	}
}

func loadPostMedia(ctx context.Context, postID uint64) ([]PostMediaDTO, error) {
	rows, err := dao.UcgPostMedia.Ctx(ctx).
		Where(dao.UcgPostMedia.Columns().PostId, postID).
		OrderAsc(dao.UcgPostMedia.Columns().SortOrder).
		All()
	if err != nil {
		return nil, err
	}
	out := make([]PostMediaDTO, 0, len(rows))
	for _, row := range rows {
		var m entity.UcgPostMedia
		if err = row.Struct(&m); err != nil {
			return nil, err
		}
		key := strings.TrimSpace(m.ObjectKey)
		cdn := BuildCdnURL(key)
		item := PostMediaDTO{
			ObjectKey: key,
			CdnUrl:    cdn,
			MediaKind: m.MediaKind,
			SortOrder: m.SortOrder,
		}
		if key != "" && m.MediaKind == 1 {
			item.ThumbnailUrl = BuildImageThumbnailURL(key)
		}
		out = append(out, item)
	}
	return out, nil
}
