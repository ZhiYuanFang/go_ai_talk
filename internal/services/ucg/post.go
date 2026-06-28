package ucg

import (
	"context"
	"strings"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/platform/eventkit"

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
	Id             uint64         `json:"id"`
	AuthorWxId     uint64         `json:"authorWxId"`
	Content        string         `json:"content"`
	Status         int            `json:"status"`
	RejectReason   string         `json:"rejectReason,omitempty"`
	MediaType      int            `json:"mediaType"`
	LikeCount      uint           `json:"likeCount"`
	CommentCount   uint           `json:"commentCount"`
	LikedByMe      bool           `json:"likedByMe"`
	CreatedAt      int64          `json:"createdAt"`
	UpdatedAt      int64          `json:"updatedAt"`
	PublishedAt    int64          `json:"publishedAt,omitempty"`
	IpLocation     string         `json:"ipLocation,omitempty"`
	DistanceMeters string         `json:"distanceMeters,omitempty"`
	Media          []PostMediaDTO `json:"media,omitempty"`
	Author         *ProfileDTO    `json:"author,omitempty"`
}

// PostMediaDTO 帖子媒体展示。
type PostMediaDTO struct {
	ObjectKey    string `json:"objectKey"`
	CdnUrl       string `json:"cdnUrl"`
	ThumbnailUrl string `json:"thumbnailUrl,omitempty"`
	MediaKind    int    `json:"mediaKind"`
	SortOrder    int    `json:"sortOrder"`
}

// CreatePost 创建帖子；submit=true 时进入 pending_audit 并事务提交后发 MQ。clientIP 用于服务端快照 ip_location。
func CreatePost(ctx context.Context, wxID int64, content string, mediaType int, submit bool, media []PostMediaInput, clientIP string, lat, lng *float64) (*PostDTO, error) {
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	status := PostStatusDraft
	auditVersion := 0
	if submit {
		status = PostStatusPendingAudit
		auditVersion = 1
	}
	now := time.Now().Unix()
	ipLoc := SnapshotPostIpLocation(ctx, clientIP)
	var postID uint64
	var outboxID uint64
	err := dao.UcgPost.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		data := g.Map{
			dao.UcgPost.Columns().AuthorWxId: wxID,
			dao.UcgPost.Columns().Content:    strings.TrimSpace(content),
			dao.UcgPost.Columns().Status:     status,
			dao.UcgPost.Columns().MediaType:  mediaType,
			dao.UcgPost.Columns().CreatedAt:  now,
			dao.UcgPost.Columns().UpdatedAt:  now,
		}
		if submit {
			data[dao.UcgPost.Columns().AuditVersion] = auditVersion
		}
		if ipLoc != "" {
			data[dao.UcgPost.Columns().IpLocation] = ipLoc
		}
		applyPostCoords(data, lat, lng)
		res, insErr := tx.Model(dao.UcgPost.Table()).Ctx(ctx).Data(data).Insert()
		if insErr != nil {
			return insErr
		}
		id, _ := res.LastInsertId()
		postID = uint64(id)
		// 插入媒体
		if insErr = replacePostMediaTx(ctx, tx, postID, media); insErr != nil {
			return insErr
		}
		// 提交审核
		if submit {
			outboxID, insErr = enqueueAuditPublishOutboxTx(ctx, tx, eventkit.RoutingUcgPostCreated.String(),
				auditPublishPostPayload(postID, auditVersion))
		}
		return insErr
	})
	if err != nil {
		return nil, err
	}
	if submit {
		scheduleAuditPublishAfterCommit(ctx, outboxID)
	}
	return GetPostByID(ctx, postID, wxID, nil, nil)
}

// UpdatePost 编辑自己的帖子；再提审递增 audit_version 并发 MQ。
func UpdatePost(ctx context.Context, wxID int64, postID uint64, content string, mediaType int, submit bool, media []PostMediaInput, lat, lng *float64) (*PostDTO, error) {
	post, err := loadOwnedPost(ctx, wxID, postID)
	if err != nil {
		return nil, err
	}
	status := post.Status
	auditVersion := post.AuditVersion
	if auditVersion <= 0 {
		auditVersion = 1
	}
	needPublish := false
	if submit {
		if post.Status == PostStatusPublished || post.Status == PostStatusRejected || post.Status == PostStatusApplyFailed {
			auditVersion++
		}
		status = PostStatusPendingAudit
		needPublish = true
	} else if status == PostStatusPublished {
		status = PostStatusPendingAudit
	}
	now := time.Now().Unix()
	finalAuditVersion := auditVersion
	var outboxID uint64
	err = dao.UcgPost.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		updateData := g.Map{
			dao.UcgPost.Columns().Content:      strings.TrimSpace(content),
			dao.UcgPost.Columns().Status:       status,
			dao.UcgPost.Columns().MediaType:    mediaType,
			dao.UcgPost.Columns().UpdatedAt:    now,
			dao.UcgPost.Columns().AuditVersion: finalAuditVersion,
		}
		// 再提审须重置机审/apply 中间态，避免沿用上轮 verdict 跳过 Green。
		if status == PostStatusPendingAudit {
			updateData[dao.UcgPost.Columns().ModerationVerdict] = ModerationVerdictNone
			updateData[dao.UcgPost.Columns().ModerationReason] = ""
			updateData[dao.UcgPost.Columns().ModerationAt] = 0
			updateData[dao.UcgPost.Columns().ApplyAttempts] = 0
			updateData[dao.UcgPost.Columns().ApplyFailedAt] = 0
			updateData[dao.UcgPost.Columns().RejectReason] = ""
		}
		if lat != nil || lng != nil {
			applyPostCoords(updateData, lat, lng)
		}
		_, uErr := tx.Model(dao.UcgPost.Table()).Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).Data(updateData).Update()
		if uErr != nil {
			return uErr
		}
		if uErr = replacePostMediaTx(ctx, tx, postID, media); uErr != nil {
			return uErr
		}
		if needPublish && status == PostStatusPendingAudit {
			outboxID, uErr = enqueueAuditPublishOutboxTx(ctx, tx, eventkit.RoutingUcgPostCreated.String(),
				auditPublishPostPayload(postID, finalAuditVersion))
		}
		return uErr
	})
	if err != nil {
		return nil, err
	}
	if needPublish && status == PostStatusPendingAudit {
		scheduleAuditPublishAfterCommit(ctx, outboxID)
	}
	dto, err := GetPostByID(ctx, postID, wxID, nil, nil)
	if err != nil {
		return nil, err
	}
	// 已发布帖坐标变更时同步 GEO/snapshot。
	if post.Status == PostStatusPublished && (lat != nil || lng != nil) {
		_ = syncPublishedPostRedis(ctx, postID)
	}
	return dto, nil
}

// DeletePost 删除自己的帖子。
func DeletePost(ctx context.Context, wxID int64, postID uint64) error {
	if _, err := loadOwnedPost(ctx, wxID, postID); err != nil {
		return err
	}
	if _, err := dao.UcgPostMedia.Ctx(ctx).Where(dao.UcgPostMedia.Columns().PostId, postID).Delete(); err != nil {
		return err
	}
	_ = RemoveRecommendScore(ctx, postID)
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

// GetPostByID 获取单帖；非 published 仅作者可见；可选 viewer 坐标返回 distanceMeters。
func GetPostByID(ctx context.Context, postID uint64, viewerWxID int64, viewerLat, viewerLng *float64) (*PostDTO, error) {
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
		liked, lErr := likedPostIDsFromRedis(ctx, viewerWxID, []uint64{postID})
		if lErr == nil {
			dto.LikedByMe = liked[postID]
		}
	}
	if dm, dErr := DistanceMetersForPost(ctx, postID, viewerLat, viewerLng); dErr == nil && dm != "" {
		dto.DistanceMeters = dm
	}
	return dto, nil
}

func applyPostCoords(data g.Map, lat, lng *float64) {
	if lat == nil && lng == nil {
		return
	}
	if lat != nil && lng != nil && validCoord(*lat, *lng) {
		data[dao.UcgPost.Columns().Lat] = *lat
		data[dao.UcgPost.Columns().Lng] = *lng
		return
	}
	data[dao.UcgPost.Columns().Lat] = nil
	data[dao.UcgPost.Columns().Lng] = nil
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
	return dao.UcgPost.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		return replacePostMediaTx(ctx, tx, postID, media)
	})
}

func replacePostMediaTx(ctx context.Context, tx gdb.TX, postID uint64, media []PostMediaInput) error {
	if _, err := tx.Model(dao.UcgPostMedia.Table()).Ctx(ctx).Where(dao.UcgPostMedia.Columns().PostId, postID).Delete(); err != nil {
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
		if _, err := tx.Model(dao.UcgPostMedia.Table()).Ctx(ctx).Data(g.Map{
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
		liked, err := likedPostIDsFromRedis(ctx, viewerWxID, postIDs)
		if err != nil {
			return nil, err
		}
		for _, dto := range out {
			dto.LikedByMe = liked[dto.Id]
		}
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
		// 视频列表缩略图用物理首帧 jpg，避免 query 截帧与 mp4 CDN 窜缓存。
		if key != "" && m.MediaKind == 2 {
			item.ThumbnailUrl = BuildVideoThumbnailURL(key)
		}
		out = append(out, item)
	}
	return out, nil
}
