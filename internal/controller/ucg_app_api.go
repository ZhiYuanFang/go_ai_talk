package controller

import (
	"context"
	"errors"
	"strconv"
	"strings"

	v1 "hello/api/v1"
	"hello/internal/services/contracts"
	"hello/internal/services/gatewayapp"
	ucgsvc "hello/internal/services/ucg"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// UcgAppCtrl UCG App HTTP API（路径前缀 /ucg/app/api；经 gateway-app 代理时由网关注入 X-Internal-Wx-Id）。
type UcgAppCtrl struct{}

// NewUcgAppCtrl 构造 UCG App 控制器。
func NewUcgAppCtrl() *UcgAppCtrl { return &UcgAppCtrl{} }

// Health GET /ucg/app/api/health
func (c *UcgAppCtrl) Health(ctx context.Context, req *v1.UcgHealthReq) (res *v1.UcgHealthRes, err error) {
	_ = c
	_ = req
	return &v1.UcgHealthRes{Status: "ok"}, nil
}

// MediaPresign POST /ucg/app/api/media/presign
func (c *UcgAppCtrl) MediaPresign(ctx context.Context, req *v1.UcgMediaPresignReq) (res *v1.UcgMediaPresignRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	uploadURL, objectKey, cdnURL, headers, err := ucgsvc.PresignUpload(ctx, wxID, req.MediaKind, req.Extension)
	if err != nil {
		return nil, err
	}
	return &v1.UcgMediaPresignRes{
		UploadUrl: uploadURL,
		ObjectKey: objectKey,
		CdnUrl:    cdnURL,
		Headers:   headers,
	}, nil
}

// MediaResolve POST /ucg/app/api/media/resolve
func (c *UcgAppCtrl) MediaResolve(ctx context.Context, req *v1.UcgMediaResolveReq) (res *v1.UcgMediaResolveRes, err error) {
	_ = c
	if _, err = wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx)); err != nil {
		return nil, err
	}
	result, err := ucgsvc.ResolveMediaByHash(ctx, req.ContentHash, req.TransformVersion, req.MediaKind)
	if err != nil {
		return nil, err
	}
	res = &v1.UcgMediaResolveRes{Hit: result.Hit}
	if result.Hit {
		res.ObjectKey = result.ObjectKey
		res.CdnUrl = result.CdnURL
	}
	return res, nil
}

// MediaRegister POST /ucg/app/api/media/register
func (c *UcgAppCtrl) MediaRegister(ctx context.Context, req *v1.UcgMediaRegisterReq) (res *v1.UcgMediaRegisterRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	result, err := ucgsvc.RegisterMedia(ctx, wxID, ucgsvc.RegisterMediaRequest{
		ObjectKey:        req.ObjectKey,
		ContentHash:      req.ContentHash,
		TransformVersion: req.TransformVersion,
		MediaKind:        req.MediaKind,
		DedupHit:         req.DedupHit,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UcgMediaRegisterRes{
		ObjectKey: result.ObjectKey,
		CdnUrl:    result.CdnURL,
	}, nil
}

// MediaDelete POST /ucg/app/api/media/delete
func (c *UcgAppCtrl) MediaDelete(ctx context.Context, req *v1.UcgMediaDeleteReq) (res *v1.UcgMediaDeleteRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	deleted, skipped, err := ucgsvc.DeleteOwnedMedia(ctx, wxID, req.ObjectKeys)
	if err != nil {
		return nil, err
	}
	return &v1.UcgMediaDeleteRes{Deleted: deleted, Skipped: skipped}, nil
}

// PostsPolish POST /ucg/app/api/posts/polish
func (c *UcgAppCtrl) PostsPolish(ctx context.Context, req *v1.UcgPostsPolishReq) (res *v1.UcgPostsPolishRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	snap, err := ucgsvc.CheckPolishAIQuota(ctx, wxID)
	if err != nil {
		if errors.Is(err, contracts.ErrAINotLoggedIn) {
			return nil, gerror.NewCode(contracts.GCodeAINotLoggedIn(), err.Error())
		}
		return nil, err
	}
	quotaDegraded := snap.Degraded && !snap.Allowed
	polished, err := ucgsvc.PolishPostText(ctx, req.ImageKeys, req.Text, quotaDegraded)
	if err != nil {
		return nil, err
	}
	if snap.Allowed {
		if _, err = ucgsvc.ConsumePolishAIQuota(ctx, wxID); err != nil {
			if errors.Is(err, contracts.ErrAIQuotaExhausted) {
				return nil, gerror.NewCode(contracts.GCodeAIQuotaExhausted(), err.Error())
			}
			return nil, err
		}
	}
	return &v1.UcgPostsPolishRes{PolishedText: polished, QuotaDegraded: quotaDegraded}, nil
}

// AIQuotaGet GET /ucg/app/api/ai-quota
func (c *UcgAppCtrl) AIQuotaGet(ctx context.Context, req *v1.UcgAppAIQuotaGetReq) (res *v1.UcgAppAIQuotaGetRes, err error) {
	_ = c
	_ = req
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	status, err := ucgsvc.GetPolishAIQuotaAppStatus(ctx, wxID)
	if err != nil {
		return nil, mapAIQuotaErr(err)
	}
	return &v1.UcgAppAIQuotaGetRes{
		Polish: v1.UcgAppAIQuotaFeatureStatus{
			Used:     status.Polish.Used,
			Limit:    status.Polish.Limit,
			Degraded: status.Polish.Degraded,
		},
	}, nil
}

func (c *UcgAppCtrl) ProfileMeGet(ctx context.Context, req *v1.UcgProfileMeGetReq) (res *v1.UcgProfileRes, err error) {
	_ = c
	_ = req
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	clientIP := ucgsvc.ClientIPFromRequest(ghttp.RequestFromCtx(ctx))
	p, err := ucgsvc.GetOrCreateMyProfile(ctx, wxID, clientIP)
	if err != nil {
		return nil, err
	}
	return profileDTOToRes(p), nil
}

// ProfileMePut PUT /ucg/app/api/profile/me 更新我的资料
func (c *UcgAppCtrl) ProfileMePut(ctx context.Context, req *v1.UcgProfileMePutReq) (res *v1.UcgProfileRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	// 更新我的资料
	p, err := ucgsvc.UpdateMyProfile(ctx, wxID, req.Nickname, req.AvatarKey, req.Bio)
	if err != nil {
		return nil, err
	}
	// 返回更新后的资料
	return profileDTOToRes(p), nil
}

func (c *UcgAppCtrl) ProfilePublicGet(ctx context.Context, req *v1.UcgProfilePublicGetReq) (res *v1.UcgProfileRes, err error) {
	_ = c
	p, err := ucgsvc.GetPublicProfile(ctx, req.WxId)
	if err != nil {
		return nil, err
	}
	res = profileDTOToRes(p)
	if viewerWxID, ok := wxIDFromUcgHeaderOptional(ghttp.RequestFromCtx(ctx)); ok && viewerWxID != int64(req.WxId) {
		if following, fErr := ucgsvc.IsFollowing(ctx, viewerWxID, int64(req.WxId)); fErr == nil {
			res.IsFollowing = following
		}
	}
	return res, nil
}

func (c *UcgAppCtrl) PostCreate(ctx context.Context, req *v1.UcgPostCreateReq) (res *v1.UcgPostCreateRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	clientIP := ucgsvc.ClientIPFromRequest(ghttp.RequestFromCtx(ctx))
	post, err := ucgsvc.CreatePostWithParams(ctx, ucgsvc.CreatePostParams{
		WxID: wxID, Content: req.Content, MediaType: req.MediaType, Submit: req.Submit,
		Media: toPostMediaInput(req.Media), ClientIP: clientIP,
		Type: req.Type, DebateLeft: req.DebateLeft, DebateRight: req.DebateRight,
		Lat: req.Lat, Lng: req.Lng,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UcgPostCreateRes{UcgPostItem: postDTOToItem(post)}, nil
}

func (c *UcgAppCtrl) PostUpdate(ctx context.Context, req *v1.UcgPostUpdateReq) (res *v1.UcgPostUpdateRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	post, err := ucgsvc.UpdatePost(ctx, wxID, req.Id, req.Content, req.MediaType, req.Submit, toPostMediaInput(req.Media), req.Lat, req.Lng)
	if err != nil {
		return nil, err
	}
	return &v1.UcgPostUpdateRes{UcgPostItem: postDTOToItem(post)}, nil
}

func (c *UcgAppCtrl) PostDelete(ctx context.Context, req *v1.UcgPostDeleteReq) (res *v1.UcgPostDeleteRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if err = ucgsvc.DeletePost(ctx, wxID, req.Id); err != nil {
		return nil, err
	}
	return &v1.UcgPostDeleteRes{}, nil
}

func (c *UcgAppCtrl) PostGet(ctx context.Context, req *v1.UcgPostGetReq) (res *v1.UcgPostGetRes, err error) {
	_ = c
	viewerWxID, _ := wxIDFromUcgHeaderOptional(ghttp.RequestFromCtx(ctx))
	post, err := ucgsvc.GetPostByID(ctx, req.Id, viewerWxID, req.Lat, req.Lng)
	if err != nil {
		return nil, err
	}
	return &v1.UcgPostGetRes{UcgPostItem: postDTOToItem(post)}, nil
}

func (c *UcgAppCtrl) PostsMine(ctx context.Context, req *v1.UcgPostsMineReq) (res *v1.UcgPageRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	page, err := ucgsvc.ListMyPosts(ctx, wxID, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return pageResultToRes(page), nil
}

func (c *UcgAppCtrl) PostsUser(ctx context.Context, req *v1.UcgPostsUserReq) (res *v1.UcgPageRes, err error) {
	_ = c
	viewerWxID, _ := wxIDFromUcgHeaderOptional(ghttp.RequestFromCtx(ctx))
	page, err := ucgsvc.ListUserPosts(ctx, int64(req.WxId), viewerWxID, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return pageResultToRes(page), nil
}

func (c *UcgAppCtrl) FeedRecommend(ctx context.Context, req *v1.UcgFeedRecommendReq) (res *v1.UcgFeedRecommendRes, err error) {
	_ = c
	viewerWxID, _ := wxIDFromUcgHeaderOptional(ghttp.RequestFromCtx(ctx))
	cursor := strings.TrimSpace(req.Cursor)
	var lat, lng *float64
	if cursor == "" {
		lat, lng = req.Lat, req.Lng
	}
	page, err := ucgsvc.ListRecommendFeed(ctx, viewerWxID, lat, lng, cursor, req.PageSize, req.Type)
	if err != nil {
		return nil, err
	}
	list := make([]v1.UcgPostItem, 0, len(page.List))
	for _, p := range page.List {
		list = append(list, postDTOToItem(p))
	}
	return &v1.UcgFeedRecommendRes{
		List:       list,
		HasMore:    page.HasMore,
		NextCursor: page.NextCursor,
	}, nil
}

func (c *UcgAppCtrl) FeedFollowing(ctx context.Context, req *v1.UcgFeedFollowingReq) (res *v1.UcgPageRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	page, err := ucgsvc.ListFollowingFeed(ctx, wxID, req.Page, req.PageSize, req.Lat, req.Lng, req.Type)
	if err != nil {
		return nil, err
	}
	return pageResultToRes(page), nil
}

func (c *UcgAppCtrl) FollowPost(ctx context.Context, req *v1.UcgFollowPostReq) (res *v1.UcgFollowDeleteRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if err = ucgsvc.Follow(ctx, wxID, int64(req.WxId)); err != nil {
		return nil, err
	}
	return &v1.UcgFollowDeleteRes{}, nil
}

func (c *UcgAppCtrl) FollowDelete(ctx context.Context, req *v1.UcgFollowDeleteReq) (res *v1.UcgFollowDeleteRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if err = ucgsvc.Unfollow(ctx, wxID, int64(req.WxId)); err != nil {
		return nil, err
	}
	return &v1.UcgFollowDeleteRes{}, nil
}

func (c *UcgAppCtrl) FollowingList(ctx context.Context, req *v1.UcgFollowingListReq) (res *v1.UcgFollowingListRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	page, err := ucgsvc.ListFollowing(ctx, wxID, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	res = &v1.UcgFollowingListRes{
		Total: page.Total, Page: page.Page, PageSize: page.PageSize,
		List: []uint64{},
	}
	if ids, ok := page.List.([]uint64); ok {
		res.List = ids
	}
	return res, nil
}

func (c *UcgAppCtrl) PostLikePost(ctx context.Context, req *v1.UcgPostLikePostReq) (res *v1.UcgPostLikePostRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if err = ucgsvc.LikePost(ctx, wxID, req.Id); err != nil {
		return nil, err
	}
	return &v1.UcgPostLikePostRes{}, nil
}

func (c *UcgAppCtrl) PostVotePost(ctx context.Context, req *v1.UcgPostVotePostReq) (res *v1.UcgPostVotePostRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if err = ucgsvc.VotePost(ctx, wxID, req.Id, req.Side); err != nil {
		return nil, err
	}
	return &v1.UcgPostVotePostRes{}, nil
}

func (c *UcgAppCtrl) PostLikeDelete(ctx context.Context, req *v1.UcgPostLikeDeleteReq) (res *v1.UcgPostLikeDeleteRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if err = ucgsvc.UnlikePost(ctx, wxID, req.Id); err != nil {
		return nil, err
	}
	return &v1.UcgPostLikeDeleteRes{}, nil
}

func (c *UcgAppCtrl) PostLikesGet(ctx context.Context, req *v1.UcgPostLikesGetReq) (res *v1.UcgLikesPageRes, err error) {
	_ = c
	page, err := ucgsvc.ListPostLikes(ctx, req.Id, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return likesPageToRes(page), nil
}

func (c *UcgAppCtrl) PostCommentsGet(ctx context.Context, req *v1.UcgPostCommentsGetReq) (res *v1.UcgCommentsListRes, err error) {
	_ = c
	viewerWxID, _ := wxIDFromUcgHeaderOptional(ghttp.RequestFromCtx(ctx))
	result, err := ucgsvc.ListComments(ctx, req.Id, viewerWxID)
	if err != nil {
		return nil, err
	}
	return commentsListToRes(result), nil
}

func (c *UcgAppCtrl) PostCommentPost(ctx context.Context, req *v1.UcgPostCommentPostReq) (res *v1.UcgPostCommentPostRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	cmt, err := ucgsvc.AddComment(ctx, wxID, req.Id, req.Content)
	if err != nil {
		return nil, err
	}
	return &v1.UcgPostCommentPostRes{UcgCommentItem: commentDTOToItem(cmt)}, nil
}

func (c *UcgAppCtrl) CommentDelete(ctx context.Context, req *v1.UcgCommentDeleteReq) (res *v1.UcgCommentDeleteRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if err = ucgsvc.DeleteComment(ctx, wxID, req.Id); err != nil {
		return nil, err
	}
	return &v1.UcgCommentDeleteRes{}, nil
}

func (c *UcgAppCtrl) CommentNotificationsGet(ctx context.Context, req *v1.UcgCommentNotificationsGetReq) (res *v1.UcgCommentNotificationsGetRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	page, err := ucgsvc.ListCommentNotifications(ctx, wxID, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return notificationPageToRes(page), nil
}

func (c *UcgAppCtrl) CommentNotificationsReadPost(ctx context.Context, req *v1.UcgCommentNotificationsReadPostReq) (res *v1.UcgCommentNotificationsReadPostRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if err = ucgsvc.MarkNotificationsRead(ctx, wxID, req.Ids, req.All); err != nil {
		return nil, err
	}
	return &v1.UcgCommentNotificationsReadPostRes{}, nil
}

func (c *UcgAppCtrl) ConversationsGet(ctx context.Context, req *v1.UcgConversationsGetReq) (res *v1.UcgConversationsPageRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	page, err := ucgsvc.ListConversations(ctx, wxID, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return conversationsPageToRes(page), nil
}

func (c *UcgAppCtrl) ConversationCreate(ctx context.Context, req *v1.UcgConversationCreateReq) (res *v1.UcgConversationCreateRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	conv, err := ucgsvc.GetOrCreateDirectConversation(ctx, wxID, req.TargetWxId)
	if err != nil {
		return nil, err
	}
	return &v1.UcgConversationCreateRes{UcgConversationItem: conversationDTOToItem(conv)}, nil
}

func (c *UcgAppCtrl) ConversationMessagesGet(ctx context.Context, req *v1.UcgConversationMessagesGetReq) (res *v1.UcgChatMessagesPageRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	page, err := ucgsvc.ListConversationMessages(ctx, wxID, req.Id, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return chatMessagesPageToRes(page), nil
}

func (c *UcgAppCtrl) ConversationReadPost(ctx context.Context, req *v1.UcgConversationReadPostReq) (res *v1.UcgConversationReadPostRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if err = ucgsvc.MarkConversationRead(ctx, wxID, req.Id, req.LastMsgId); err != nil {
		return nil, err
	}
	return &v1.UcgConversationReadPostRes{}, nil
}

func (c *UcgAppCtrl) ConversationPinPut(ctx context.Context, req *v1.UcgConversationPinPutReq) (res *v1.UcgConversationPinPutRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if err = ucgsvc.SetConversationPinned(ctx, wxID, req.Id, req.Pinned); err != nil {
		return nil, err
	}
	return &v1.UcgConversationPinPutRes{}, nil
}

func (c *UcgAppCtrl) ConversationDelete(ctx context.Context, req *v1.UcgConversationDeleteReq) (res *v1.UcgConversationDeleteRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if err = ucgsvc.SoftDeleteConversation(ctx, wxID, req.Id); err != nil {
		return nil, err
	}
	return &v1.UcgConversationDeleteRes{}, nil
}

func (c *UcgAppCtrl) PushRegisterPost(ctx context.Context, req *v1.UcgPushRegisterPostReq) (res *v1.UcgPushRegisterPostRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if err = ucgsvc.RegisterPushDevice(ctx, wxID, req.Channel, req.Token, req.DeviceKey); err != nil {
		return nil, err
	}
	return &v1.UcgPushRegisterPostRes{}, nil
}

func (c *UcgAppCtrl) PushUnregisterPost(ctx context.Context, req *v1.UcgPushUnregisterPostReq) (res *v1.UcgPushUnregisterPostRes, err error) {
	_ = c
	wxID, err := wxIDFromUcgHeader(ghttp.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if err = ucgsvc.UnregisterPushDevice(ctx, wxID, req.DeviceKey, req.Channel); err != nil {
		return nil, err
	}
	return &v1.UcgPushUnregisterPostRes{}, nil
}

func conversationsPageToRes(page *ucgsvc.PageResult) *v1.UcgConversationsPageRes {
	res := &v1.UcgConversationsPageRes{
		Total: page.Total, Page: page.Page, PageSize: page.PageSize,
		List: []v1.UcgConversationItem{},
	}
	if items, ok := page.List.([]*ucgsvc.ConversationDTO); ok {
		for _, it := range items {
			res.List = append(res.List, conversationDTOToItem(it))
		}
	}
	return res
}

func conversationDTOToItem(c *ucgsvc.ConversationDTO) v1.UcgConversationItem {
	if c == nil {
		return v1.UcgConversationItem{}
	}
	return v1.UcgConversationItem{
		Id: c.Id, PeerWxId: c.PeerWxId, Pinned: c.Pinned,
		UnreadCount: c.UnreadCount, UpdatedAt: c.UpdatedAt, LastPreview: c.LastPreview,
		PeerNickname: c.PeerNickname, PeerAvatarKey: c.PeerAvatarKey, PeerAvatarUrl: c.PeerAvatarUrl,
		PeerAvatarThumbnailUrl: c.PeerAvatarThumbnailUrl,
	}
}

func chatMessagesPageToRes(page *ucgsvc.PageResult) *v1.UcgChatMessagesPageRes {
	res := &v1.UcgChatMessagesPageRes{
		Total: page.Total, Page: page.Page, PageSize: page.PageSize,
		List: []v1.UcgChatMessageItem{},
	}
	if msgs, ok := page.List.([]ucgsvc.ChatMessage); ok {
		for _, m := range msgs {
			res.List = append(res.List, v1.UcgChatMessageItem{
				Id: m.ID, ClientMsgId: m.ClientMsgID, SenderWxId: m.SenderWxID,
				Content: m.Content, CreatedAt: m.CreatedAt, Status: m.Status,
				ImageKey: m.ImageKey, VideoKey: m.VideoKey,
				MediaCdnUrl: m.MediaCdnUrl, MediaThumbnailUrl: m.MediaThumbnailUrl,
			})
		}
	}
	return res
}

func likesPageToRes(page *ucgsvc.PageResult) *v1.UcgLikesPageRes {
	res := &v1.UcgLikesPageRes{
		Total: page.Total, Page: page.Page, PageSize: page.PageSize,
		List: []v1.UcgLikerItem{},
	}
	if likers, ok := page.List.([]*ucgsvc.LikerDTO); ok {
		for _, l := range likers {
			if l == nil {
				continue
			}
			res.List = append(res.List, v1.UcgLikerItem{
				WxId: l.WxId, Nickname: l.Nickname,
				AvatarKey: l.AvatarKey, AvatarUrl: l.AvatarUrl,
				AvatarThumbnailUrl: l.AvatarThumbnailUrl,
			})
		}
	}
	return res
}

func commentsListToRes(result *ucgsvc.CommentsListResult) *v1.UcgCommentsListRes {
	res := &v1.UcgCommentsListRes{
		Total: result.Total, Truncated: result.Truncated,
		List: []v1.UcgCommentItem{},
	}
	for _, c := range result.List {
		res.List = append(res.List, commentDTOToItem(c))
	}
	return res
}

func notificationPageToRes(page *ucgsvc.NotificationPageResult) *v1.UcgCommentNotificationsGetRes {
	if page == nil {
		return &v1.UcgCommentNotificationsGetRes{List: []v1.UcgCommentNotificationItem{}}
	}
	res := &v1.UcgCommentNotificationsGetRes{
		Total: page.Total, Page: page.Page, PageSize: page.PageSize,
		UnreadCount: page.UnreadCount,
		List:        []v1.UcgCommentNotificationItem{},
	}
	for _, n := range page.List {
		res.List = append(res.List, notificationDTOToItem(n))
	}
	return res
}

func notificationDTOToItem(n *ucgsvc.NotificationDTO) v1.UcgCommentNotificationItem {
	if n == nil {
		return v1.UcgCommentNotificationItem{}
	}
	item := v1.UcgCommentNotificationItem{
		Id: n.Id, Type: n.Type, PostId: n.PostId, CommentId: n.CommentId,
		Preview: n.Preview, PostThumbUrl: n.PostThumbUrl, PostMediaKind: n.PostMediaKind,
		Read: n.Read, CreatedAt: n.CreatedAt,
	}
	if n.Actor != nil {
		item.Actor = profileDTOToRes(n.Actor)
	}
	return item
}

func commentDTOToItem(c *ucgsvc.CommentDTO) v1.UcgCommentItem {
	if c == nil {
		return v1.UcgCommentItem{}
	}
	item := v1.UcgCommentItem{
		Id: c.Id, PostId: c.PostId, AuthorWxId: c.AuthorWxId,
		Content: c.Content, CreatedAt: c.CreatedAt,
		Status: c.Status, RejectReason: c.RejectReason, AuditVersion: c.AuditVersion,
		VoteSide: c.VoteSide, VoteSideLabel: c.VoteSideLabel,
	}
	if c.Author != nil {
		item.Author = profileDTOToRes(c.Author)
	}
	return item
}

func wxIDFromUcgHeader(r *ghttp.Request) (int64, error) {
	wxID, ok := wxIDFromUcgHeaderOptional(r)
	if !ok {
		return 0, gerror.NewCode(gcode.CodeInvalidParameter, "缺少 X-Internal-Wx-Id")
	}
	return wxID, nil
}

// wxIDFromUcgHeaderOptional 解析网关注入的 wxId；缺失或无效时返回 (0, false)，供匿名可读接口可选身份。
func wxIDFromUcgHeaderOptional(r *ghttp.Request) (int64, bool) {
	if r == nil {
		return 0, false
	}
	s := strings.TrimSpace(r.GetHeader(gatewayapp.HeaderInternalWxId))
	if s == "" {
		return 0, false
	}
	wxID, err := strconv.ParseInt(s, 10, 64)
	if err != nil || wxID <= 0 {
		return 0, false
	}
	return wxID, true
}

// profileDTOToRes 将 ucgsvc.ProfileDTO 转换为 v1.UcgProfileRes
func profileDTOToRes(p *ucgsvc.ProfileDTO) *v1.UcgProfileRes {
	if p == nil {
		return nil
	}
	return &v1.UcgProfileRes{
		WxId:               p.WxId,
		Nickname:           p.Nickname,
		AvatarKey:          p.AvatarKey,
		AvatarUrl:          p.AvatarUrl,
		AvatarThumbnailUrl: p.AvatarThumbnailUrl,
		Bio:                p.Bio,
		FollowerCount:      p.FollowerCount,
		FollowingCount:     p.FollowingCount,
		PostCount:          p.PostCount,
		IpLocation:         p.IpLocation,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
		AuditPending:       p.AuditPending,
		RejectReason:       p.RejectReason,
		ForceValue:         p.ForceValue,
		ForceTier:          p.ForceTier,
	}
}

func toPostMediaInput(in []v1.UcgPostMediaInput) []ucgsvc.PostMediaInput {
	out := make([]ucgsvc.PostMediaInput, 0, len(in))
	for _, m := range in {
		out = append(out, ucgsvc.PostMediaInput{
			ObjectKey:  m.ObjectKey,
			MediaKind:  m.MediaKind,
			SortOrder:  m.SortOrder,
			DurationMs: m.DurationMs,
			SizeBytes:  m.SizeBytes,
		})
	}
	return out
}

func postDTOToItem(p *ucgsvc.PostDTO) v1.UcgPostItem {
	if p == nil {
		return v1.UcgPostItem{}
	}
	media := make([]v1.UcgPostMediaOut, 0, len(p.Media))
	for _, m := range p.Media {
		media = append(media, v1.UcgPostMediaOut{
			ObjectKey:    m.ObjectKey,
			CdnUrl:       m.CdnUrl,
			ThumbnailUrl: m.ThumbnailUrl,
			MediaKind:    m.MediaKind,
			SortOrder:    m.SortOrder,
		})
	}
	item := v1.UcgPostItem{
		Id:             p.Id,
		AuthorWxId:     p.AuthorWxId,
		Type:           p.Type,
		Content:        p.Content,
		DebateLeft:     p.DebateLeft,
		DebateRight:    p.DebateRight,
		Status:         p.Status,
		RejectReason:   p.RejectReason,
		MediaType:      p.MediaType,
		LikeCount:      p.LikeCount,
		CommentCount:   p.CommentCount,
		LeftVoteCount:  p.LeftVoteCount,
		RightVoteCount: p.RightVoteCount,
		MyVoteSide:     p.MyVoteSide,
		LikedByMe:      p.LikedByMe,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		PublishedAt:    p.PublishedAt,
		IpLocation:     p.IpLocation,
		DistanceMeters: p.DistanceMeters,
		Media:          media,
	}
	if p.Author != nil {
		item.Author = profileDTOToRes(p.Author)
	}
	return item
}

func pageResultToRes(page *ucgsvc.PageResult) *v1.UcgPageRes {
	res := &v1.UcgPageRes{
		Total:    page.Total,
		Page:     page.Page,
		PageSize: page.PageSize,
		List:     []v1.UcgPostItem{},
	}
	if posts, ok := page.List.([]*ucgsvc.PostDTO); ok {
		for _, p := range posts {
			res.List = append(res.List, postDTOToItem(p))
		}
	}
	return res
}
