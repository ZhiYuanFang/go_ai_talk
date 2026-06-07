package v1

import "github.com/gogf/gf/v2/frame/g"

// UcgMediaPresignReq OSS 直传预签名（objectKey 由服务端生成，客户端不可指定 bucket）。
type UcgMediaPresignReq struct {
	g.Meta    `path:"/ucg/app/api/media/presign" method:"post" tags:"ucg" summary:"获取 OSS 上传预签名"`
	MediaKind int    `json:"mediaKind" v:"required|in:1,2" dc:"1=图片 2=视频"`
	Extension string `json:"extension" v:"required|length:1,16" dc:"文件扩展名，不含点，如 jpg mp4"`
}

// UcgMediaPresignRes 预签名响应；DB 仅存 objectKey，展示用 cdnUrl。
type UcgMediaPresignRes struct {
	UploadUrl string            `json:"uploadUrl"`
	ObjectKey string            `json:"objectKey"`
	CdnUrl    string            `json:"cdnUrl"`
	Headers   map[string]string `json:"headers"`
}

// UcgHealthReq 进程健康探测（供 compose/k8s 与联调）。
type UcgHealthReq struct {
	g.Meta `path:"/ucg/app/api/health" method:"get" tags:"ucg" summary:"UCG 服务健康检查"`
}

// UcgHealthRes 健康检查响应。
type UcgHealthRes struct {
	Status string `json:"status"`
}

type UcgProfileMeGetReq struct {
	g.Meta `path:"/ucg/app/api/profile/me" method:"get" tags:"ucg" summary:"当前用户 profile"`
}

type UcgProfileMePutReq struct {
	g.Meta     `path:"/ucg/app/api/profile/me" method:"put" tags:"ucg" summary:"更新 profile"`
	Nickname   string `json:"nickname"`
	AvatarKey  string `json:"avatarKey"`
	Bio        string `json:"bio"`
}

type UcgProfilePublicGetReq struct {
	g.Meta `path:"/ucg/app/api/profile/{wxId}" method:"get" tags:"ucg" summary:"公开 profile"`
	WxId   uint64 `json:"wxId" in:"path" v:"required|min:1"`
}

type UcgProfileRes struct {
	WxId         uint64 `json:"wxId"`
	Nickname     string `json:"nickname"`
	AvatarKey    string `json:"avatarKey"`
	AvatarUrl    string `json:"avatarUrl"`
	Bio          string `json:"bio"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
	AuditPending bool   `json:"auditPending,omitempty"`
	RejectReason string `json:"rejectReason,omitempty"`
}

type UcgPostCreateReq struct {
	g.Meta    `path:"/ucg/app/api/posts" method:"post" tags:"ucg" summary:"创建帖子"`
	Content   string              `json:"content"`
	MediaType int                 `json:"mediaType"`
	Submit    bool                `json:"submit"`
	Media     []UcgPostMediaInput `json:"media"`
}

type UcgPostUpdateReq struct {
	g.Meta    `path:"/ucg/app/api/posts/{id}" method:"put" tags:"ucg" summary:"更新帖子"`
	Id        uint64              `json:"id" in:"path" v:"required|min:1"`
	Content   string              `json:"content"`
	MediaType int                 `json:"mediaType"`
	Submit    bool                `json:"submit"`
	Media     []UcgPostMediaInput `json:"media"`
}

type UcgPostDeleteReq struct {
	g.Meta `path:"/ucg/app/api/posts/{id}" method:"delete" tags:"ucg" summary:"删除帖子"`
	Id     uint64 `json:"id" in:"path" v:"required|min:1"`
}

type UcgPostDeleteRes struct{}

// UcgPostCreateRes 创建帖子响应（字段与 UcgPostItem 一致）。
type UcgPostCreateRes struct {
	UcgPostItem
}

// UcgPostUpdateRes 更新帖子响应。
type UcgPostUpdateRes struct {
	UcgPostItem
}

type UcgPostsMineReq struct {
	g.Meta   `path:"/ucg/app/api/posts/mine" method:"get" tags:"ucg" summary:"我的动态"`
	Page     int `json:"page" in:"query" d:"1"`
	PageSize int `json:"pageSize" in:"query" d:"20"`
}

type UcgFeedRecommendReq struct {
	g.Meta   `path:"/ucg/app/api/feed/recommend" method:"get" tags:"ucg" summary:"推荐 Feed"`
	Page     int `json:"page" in:"query" d:"1"`
	PageSize int `json:"pageSize" in:"query" d:"20"`
}

type UcgFeedFollowingReq struct {
	g.Meta   `path:"/ucg/app/api/feed/following" method:"get" tags:"ucg" summary:"关注 Feed"`
	Page     int `json:"page" in:"query" d:"1"`
	PageSize int `json:"pageSize" in:"query" d:"20"`
}

type UcgPostMediaInput struct {
	ObjectKey  string `json:"objectKey"`
	MediaKind  int    `json:"mediaKind"`
	SortOrder  int    `json:"sortOrder"`
	DurationMs int    `json:"durationMs"`
	SizeBytes  int64  `json:"sizeBytes"`
}

type UcgPostItem struct {
	Id           uint64            `json:"id"`
	AuthorWxId   uint64            `json:"authorWxId"`
	Content      string            `json:"content"`
	Status       int               `json:"status"`
	RejectReason string            `json:"rejectReason,omitempty"`
	MediaType    int               `json:"mediaType"`
	LikeCount    uint              `json:"likeCount"`
	CommentCount uint              `json:"commentCount"`
	LikedByMe    bool              `json:"likedByMe"`
	CreatedAt    int64             `json:"createdAt"`
	UpdatedAt    int64             `json:"updatedAt"`
	PublishedAt  int64             `json:"publishedAt,omitempty"`
	Media        []UcgPostMediaOut `json:"media,omitempty"`
	Author       *UcgProfileRes    `json:"author,omitempty"`
}

type UcgPostMediaOut struct {
	ObjectKey string `json:"objectKey"`
	CdnUrl    string `json:"cdnUrl"`
	MediaKind int    `json:"mediaKind"`
	SortOrder int    `json:"sortOrder"`
}

type UcgPageRes struct {
	List     []UcgPostItem `json:"list"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

type UcgFollowPostReq struct {
	g.Meta `path:"/ucg/app/api/follow/{wxId}" method:"post" tags:"ucg" summary:"关注用户"`
	WxId   uint64 `json:"wxId" in:"path" v:"required|min:1"`
}

type UcgFollowDeleteReq struct {
	g.Meta `path:"/ucg/app/api/follow/{wxId}" method:"delete" tags:"ucg" summary:"取消关注"`
	WxId   uint64 `json:"wxId" in:"path" v:"required|min:1"`
}

type UcgFollowDeleteRes struct{}

type UcgFollowingListReq struct {
	g.Meta   `path:"/ucg/app/api/follow/following" method:"get" tags:"ucg" summary:"我关注的人"`
	Page     int `json:"page" in:"query" d:"1"`
	PageSize int `json:"pageSize" in:"query" d:"20"`
}

type UcgFollowingListRes struct {
	List     []uint64 `json:"list"`
	Total    int      `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"pageSize"`
}

type UcgPostLikePostReq struct {
	g.Meta `path:"/ucg/app/api/posts/{id}/like" method:"post" tags:"ucg" summary:"点赞"`
	Id     uint64 `json:"id" in:"path" v:"required|min:1"`
}

type UcgPostLikePostRes struct{}

type UcgPostLikeDeleteReq struct {
	g.Meta `path:"/ucg/app/api/posts/{id}/like" method:"delete" tags:"ucg" summary:"取消赞"`
	Id     uint64 `json:"id" in:"path" v:"required|min:1"`
}

type UcgPostLikeDeleteRes struct{}

type UcgPostLikesGetReq struct {
	g.Meta   `path:"/ucg/app/api/posts/{id}/likes" method:"get" tags:"ucg" summary:"点赞用户列表"`
	Id       uint64 `json:"id" in:"path" v:"required|min:1"`
	Page     int    `json:"page" in:"query" d:"1"`
	PageSize int    `json:"pageSize" in:"query" d:"20"`
}

type UcgLikerItem struct {
	WxId     uint64 `json:"wxId"`
	Nickname string `json:"nickname"`
}

type UcgLikesPageRes struct {
	List     []UcgLikerItem `json:"list"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

type UcgPostCommentsGetReq struct {
	g.Meta   `path:"/ucg/app/api/posts/{id}/comments" method:"get" tags:"ucg" summary:"评论列表"`
	Id       uint64 `json:"id" in:"path" v:"required|min:1"`
	Page     int    `json:"page" in:"query" d:"1"`
	PageSize int    `json:"pageSize" in:"query" d:"20"`
}

type UcgPostCommentPostReq struct {
	g.Meta  `path:"/ucg/app/api/posts/{id}/comments" method:"post" tags:"ucg" summary:"发表评论"`
	Id      uint64 `json:"id" in:"path" v:"required|min:1"`
	Content string `json:"content" v:"required"`
}

// UcgPostCommentPostRes 发表评论响应。
type UcgPostCommentPostRes struct {
	UcgCommentItem
}

type UcgCommentItem struct {
	Id         uint64         `json:"id"`
	PostId     uint64         `json:"postId"`
	AuthorWxId uint64         `json:"authorWxId"`
	Content    string         `json:"content"`
	CreatedAt  int64          `json:"createdAt"`
	Author     *UcgProfileRes `json:"author,omitempty"`
}

type UcgCommentsPageRes struct {
	List     []UcgCommentItem `json:"list"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

type UcgCommentDeleteReq struct {
	g.Meta `path:"/ucg/app/api/comments/{id}" method:"delete" tags:"ucg" summary:"删除评论"`
	Id     uint64 `json:"id" in:"path" v:"required|min:1"`
}

type UcgCommentDeleteRes struct{}

type UcgConversationsGetReq struct {
	g.Meta   `path:"/ucg/app/api/conversations" method:"get" tags:"ucg" summary:"会话列表"`
	Page     int `json:"page" in:"query" d:"1"`
	PageSize int `json:"pageSize" in:"query" d:"20"`
}

type UcgConversationCreateReq struct {
	g.Meta      `path:"/ucg/app/api/conversations" method:"post" tags:"ucg" summary:"创建或获取 1:1 会话"`
	TargetWxId  int64 `json:"targetWxId" v:"required|min:1"`
}

// UcgConversationCreateRes 创建会话响应。
type UcgConversationCreateRes struct {
	UcgConversationItem
}

type UcgConversationItem struct {
	Id          uint64 `json:"id"`
	PeerWxId    uint64 `json:"peerWxId"`
	Pinned      int    `json:"pinned"`
	UnreadCount int    `json:"unreadCount"`
	UpdatedAt   int64  `json:"updatedAt"`
	LastPreview string `json:"lastPreview,omitempty"`
}

type UcgConversationsPageRes struct {
	List     []UcgConversationItem `json:"list"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"pageSize"`
}

type UcgConversationMessagesGetReq struct {
	g.Meta   `path:"/ucg/app/api/conversations/{id}/messages" method:"get" tags:"ucg" summary:"会话消息"`
	Id       uint64 `json:"id" in:"path" v:"required|min:1"`
	Page     int    `json:"page" in:"query" d:"1"`
	PageSize int    `json:"pageSize" in:"query" d:"20"`
}

type UcgChatMessageItem struct {
	Id          uint64 `json:"id"`
	ClientMsgId string `json:"clientMsgId,omitempty"`
	SenderWxId  int64  `json:"senderWxId"`
	Content     string `json:"content"`
	CreatedAt   int64  `json:"createdAt"`
	Status      string `json:"status"`
}

type UcgChatMessagesPageRes struct {
	List     []UcgChatMessageItem `json:"list"`
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
}

type UcgConversationReadPostReq struct {
	g.Meta      `path:"/ucg/app/api/conversations/{id}/read" method:"post" tags:"ucg" summary:"标记已读"`
	Id          uint64 `json:"id" in:"path" v:"required|min:1"`
	LastMsgId   uint64 `json:"lastMsgId"`
}

type UcgConversationReadPostRes struct{}

type UcgConversationPinPutReq struct {
	g.Meta `path:"/ucg/app/api/conversations/{id}/pin" method:"put" tags:"ucg" summary:"置顶会话"`
	Id     uint64 `json:"id" in:"path" v:"required|min:1"`
	Pinned bool   `json:"pinned"`
}

type UcgConversationPinPutRes struct{}

type UcgConversationDeleteReq struct {
	g.Meta `path:"/ucg/app/api/conversations/{id}" method:"delete" tags:"ucg" summary:"删除会话（软删）"`
	Id     uint64 `json:"id" in:"path" v:"required|min:1"`
}

type UcgConversationDeleteRes struct{}
