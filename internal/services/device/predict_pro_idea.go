package device

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const (
	predictProIdeaMaxTextLen  = 200
	predictProIdeaDailyLimit  = 10
	predictProIdeaDefaultPage = 20
	predictProIdeaMaxPageSize = 100
	predictProIdeaTable       = "predict_pro_idea"
	predictProIdeaLikeTable   = "predict_pro_idea_like"
	// predictProIdeaAliveWhere 未删谓词：兼容线上默认 0 与存量 NULL（>0 为软删 unix 秒）。
	predictProIdeaAliveWhere = "(deleted_at IS NULL OR deleted_at = 0)"
)

var (
	// ErrPredictProIdeaNotFound 想法不存在或已软删。
	ErrPredictProIdeaNotFound = errors.New("想法不存在")
	// ErrPredictProIdeaForbidden 非作者不得删除他人想法。
	ErrPredictProIdeaForbidden = errors.New("无权删除该想法")
	// ErrPredictProIdeaQuota 当日提交次数已达上限。
	ErrPredictProIdeaQuota = errors.New("今日提交次数已达上限")
	// ErrPredictProIdeaNoBind 未绑定设备，无法提交。
	ErrPredictProIdeaNoBind = errors.New("请先绑定设备后再提交想法")
	// ErrPredictProIdeaNoBabyName 宝宝名为空，无法提交。
	ErrPredictProIdeaNoBabyName = errors.New("请先完善宝宝昵称后再提交想法")
)

// PredictProIdeaAppItem App 列表项（不下发 wxId/deviceNo）。
type PredictProIdeaAppItem struct {
	Id         int64  `json:"id"`
	Text       string `json:"text"`
	BabyName   string `json:"babyName"`
	LikeCount  int    `json:"likeCount"`
	LikedByMe  bool   `json:"likedByMe"`
	IsMine     bool   `json:"isMine"`
	CreatedAt  int64  `json:"createdAt"`
}

// PredictProIdeaAdminItem Admin 列表项。
type PredictProIdeaAdminItem struct {
	Id          int64  `json:"id"`
	WxId        int64  `json:"wxId"`
	BabyName    string `json:"babyName"`
	Content     string `json:"content"`
	DisplayText string `json:"displayText"`
	LikeCount   int    `json:"likeCount"`
	CreatedAt   int64  `json:"createdAt"`
}

// PredictProIdeaAdminPage Admin 分页结果。
type PredictProIdeaAdminPage struct {
	List     []PredictProIdeaAdminItem
	Total    int
	Page     int
	PageSize int
}

// PredictProIdeaLikeResult 点赞/取消后热度。
type PredictProIdeaLikeResult struct {
	LikeCount int  `json:"likeCount"`
	LikedByMe bool `json:"likedByMe"`
}

type predictProIdeaRow struct {
	Id          int64  `json:"id"`
	WxId        int64  `json:"wx_id"`
	DeviceNo    string `json:"device_no"`
	BabyName    string `json:"baby_name"`
	Content     string `json:"content"`
	DisplayText string `json:"display_text"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	DeletedAt   *int64 `json:"deleted_at"`
}

func validatePredictProIdeaText(text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errors.New("想法不能为空")
	}
	if utf8.RuneCountInString(text) > predictProIdeaMaxTextLen {
		return "", errors.New("想法不能超过 200 字")
	}
	return text, nil
}

func normalizePredictProIdeaPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = predictProIdeaDefaultPage
	}
	if pageSize > predictProIdeaMaxPageSize {
		pageSize = predictProIdeaMaxPageSize
	}
	return page, pageSize
}

// shanghaiDayRangeUnix 返回 Asia/Shanghai 自然日 [start, end) 的 unix 秒。
func shanghaiDayRangeUnix(now time.Time) (start, end int64, err error) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return 0, 0, err
	}
	local := now.In(loc)
	y, m, d := local.Date()
	dayStart := time.Date(y, m, d, 0, 0, 0, 0, loc)
	return dayStart.Unix(), dayStart.Add(24 * time.Hour).Unix(), nil
}

// countPredictProIdeaToday 当日提交数（含软删，防删后再刷）。
func countPredictProIdeaToday(ctx context.Context, wxID int64) (int, error) {
	start, end, err := shanghaiDayRangeUnix(time.Now())
	if err != nil {
		return 0, err
	}
	n, err := g.DB().Model(predictProIdeaTable).Ctx(ctx).
		Where("wx_id", wxID).
		WhereGTE("created_at", start).
		WhereLT("created_at", end).
		Count()
	return n, err
}

func countIdeaLikes(ctx context.Context, ideaID int64) (int, error) {
	n, err := g.DB().Model(predictProIdeaLikeTable).Ctx(ctx).Where("idea_id", ideaID).Count()
	return n, err
}

func hasIdeaLike(ctx context.Context, ideaID, wxID int64) (bool, error) {
	one, err := g.DB().Model(predictProIdeaLikeTable).Ctx(ctx).
		Where("idea_id", ideaID).Where("wx_id", wxID).One()
	if err != nil {
		return false, err
	}
	return !one.IsEmpty(), nil
}

func loadActiveIdea(ctx context.Context, ideaID int64) (*predictProIdeaRow, error) {
	// 业务逻辑：仅加载未删想法（deleted_at=0 或 NULL）。
	one, err := g.DB().Model(predictProIdeaTable).Ctx(ctx).
		Where("id", ideaID).Where(predictProIdeaAliveWhere).One()
	if err != nil {
		return nil, err
	}
	if one.IsEmpty() {
		return nil, ErrPredictProIdeaNotFound
	}
	var row predictProIdeaRow
	if err := one.Struct(&row); err != nil {
		return nil, err
	}
	return &row, nil
}

// ListPredictProIdeasApp App 公开列表：未删除、新在前。
// Args: viewerWxID 当前登录用户（算 likedByMe/isMine）。
func (s *service) ListPredictProIdeasApp(ctx context.Context, viewerWxID int64) ([]PredictProIdeaAppItem, error) {
	var rows []predictProIdeaRow
	// 业务逻辑：列表必须包含 deleted_at=0 的活跃行，不能只用 IS NULL。
	err := g.DB().Model(predictProIdeaTable).Ctx(ctx).
		Where(predictProIdeaAliveWhere).
		OrderDesc("created_at").OrderDesc("id").
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	out := make([]PredictProIdeaAppItem, 0, len(rows))
	for _, row := range rows {
		likeCount, err := countIdeaLikes(ctx, row.Id)
		if err != nil {
			return nil, err
		}
		liked := false
		if viewerWxID > 0 {
			liked, err = hasIdeaLike(ctx, row.Id, viewerWxID)
			if err != nil {
				return nil, err
			}
		}
		out = append(out, PredictProIdeaAppItem{
			Id:        row.Id,
			Text:      row.DisplayText,
			BabyName:  row.BabyName,
			LikeCount: likeCount,
			LikedByMe: liked,
			IsMine:    viewerWxID > 0 && row.WxId == viewerWxID,
			CreatedAt: row.CreatedAt,
		})
	}
	return out, nil
}

// SubmitPredictProIdea 提交想法：须绑定设备与宝宝名；日限额 COUNT；不查 Pro 权益。
func (s *service) SubmitPredictProIdea(ctx context.Context, wxID int64, text string) (*PredictProIdeaAppItem, error) {
	if wxID <= 0 {
		return nil, errors.New("wxId 无效")
	}
	text, err := validatePredictProIdeaText(text)
	if err != nil {
		return nil, err
	}
	deviceNo, err := WxDeviceNoByWxID(ctx, wxID)
	if err != nil || strings.TrimSpace(deviceNo) == "" {
		return nil, ErrPredictProIdeaNoBind
	}
	deviceNo = strings.TrimSpace(deviceNo)
	babyName, err := ucgBabyNameByDeviceNo(ctx, deviceNo)
	if err != nil {
		return nil, err
	}
	babyName = strings.TrimSpace(babyName)
	if babyName == "" {
		return nil, ErrPredictProIdeaNoBabyName
	}
	used, err := countPredictProIdeaToday(ctx, wxID)
	if err != nil {
		return nil, err
	}
	if used >= predictProIdeaDailyLimit {
		return nil, ErrPredictProIdeaQuota
	}
	now := time.Now().Unix()
	// 业务逻辑：显式写入 deleted_at=0，避免环境 DEFAULT 为 NULL 时与「0=未删」分裂。
	id, err := g.DB().Model(predictProIdeaTable).Ctx(ctx).Data(g.Map{
		"wx_id":        wxID,
		"device_no":    deviceNo,
		"baby_name":    babyName,
		"content":      text,
		"display_text": text,
		"created_at":   now,
		"updated_at":   now,
		"deleted_at":   0,
	}).InsertAndGetId()
	if err != nil {
		return nil, err
	}
	glog.Infof(ctx, "[predict-pro-idea] submit ok wxId=%d id=%d", wxID, id)
	return &PredictProIdeaAppItem{
		Id:        id,
		Text:      text,
		BabyName:  babyName,
		LikeCount: 0,
		LikedByMe: false,
		IsMine:    true,
		CreatedAt: now,
	}, nil
}

// LikePredictProIdea 点赞（幂等插入）。
func (s *service) LikePredictProIdea(ctx context.Context, wxID, ideaID int64) (*PredictProIdeaLikeResult, error) {
	if wxID <= 0 || ideaID <= 0 {
		return nil, errors.New("参数无效")
	}
	if _, err := loadActiveIdea(ctx, ideaID); err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	_, err := g.DB().Model(predictProIdeaLikeTable).Ctx(ctx).Data(g.Map{
		"idea_id":    ideaID,
		"wx_id":      wxID,
		"created_at": now,
	}).InsertIgnore()
	if err != nil {
		return nil, err
	}
	return buildLikeResult(ctx, ideaID, wxID)
}

// UnlikePredictProIdea 取消点赞（幂等删除）。
func (s *service) UnlikePredictProIdea(ctx context.Context, wxID, ideaID int64) (*PredictProIdeaLikeResult, error) {
	if wxID <= 0 || ideaID <= 0 {
		return nil, errors.New("参数无效")
	}
	if _, err := loadActiveIdea(ctx, ideaID); err != nil {
		return nil, err
	}
	_, err := g.DB().Model(predictProIdeaLikeTable).Ctx(ctx).
		Where("idea_id", ideaID).Where("wx_id", wxID).Delete()
	if err != nil {
		return nil, err
	}
	return buildLikeResult(ctx, ideaID, wxID)
}

func buildLikeResult(ctx context.Context, ideaID, wxID int64) (*PredictProIdeaLikeResult, error) {
	n, err := countIdeaLikes(ctx, ideaID)
	if err != nil {
		return nil, err
	}
	liked, err := hasIdeaLike(ctx, ideaID, wxID)
	if err != nil {
		return nil, err
	}
	return &PredictProIdeaLikeResult{LikeCount: n, LikedByMe: liked}, nil
}

// ListPredictProIdeasAdmin Admin 分页（默认排除软删）。
func (s *service) ListPredictProIdeasAdmin(ctx context.Context, page, pageSize int) (*PredictProIdeaAdminPage, error) {
	page, pageSize = normalizePredictProIdeaPage(page, pageSize)
	// 业务逻辑：与 App 同一未删谓词，避免 Hub 空而库中有 deleted_at=0 行。
	m := g.DB().Model(predictProIdeaTable).Ctx(ctx).Where(predictProIdeaAliveWhere)
	total, err := m.Count()
	if err != nil {
		return nil, err
	}
	var rows []predictProIdeaRow
	err = g.DB().Model(predictProIdeaTable).Ctx(ctx).
		Where(predictProIdeaAliveWhere).
		OrderDesc("created_at").OrderDesc("id").
		Page(page, pageSize).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	list := make([]PredictProIdeaAdminItem, 0, len(rows))
	for _, row := range rows {
		n, err := countIdeaLikes(ctx, row.Id)
		if err != nil {
			return nil, err
		}
		list = append(list, PredictProIdeaAdminItem{
			Id:          row.Id,
			WxId:        row.WxId,
			BabyName:    row.BabyName,
			Content:     row.Content,
			DisplayText: row.DisplayText,
			LikeCount:   n,
			CreatedAt:   row.CreatedAt,
		})
	}
	return &PredictProIdeaAdminPage{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

// UpdatePredictProIdeaDisplayText Admin 改展示文案（不改 content）。
func (s *service) UpdatePredictProIdeaDisplayText(ctx context.Context, ideaID int64, displayText string) error {
	displayText, err := validatePredictProIdeaText(displayText)
	if err != nil {
		return err
	}
	if _, err := loadActiveIdea(ctx, ideaID); err != nil {
		return err
	}
	now := time.Now().Unix()
	_, err = g.DB().Model(predictProIdeaTable).Ctx(ctx).
		Where("id", ideaID).Where(predictProIdeaAliveWhere).
		Data(g.Map{"display_text": displayText, "updated_at": now}).
		Update()
	return err
}

// SoftDeletePredictProIdea Admin 软删。
func (s *service) SoftDeletePredictProIdea(ctx context.Context, ideaID int64) error {
	if _, err := loadActiveIdea(ctx, ideaID); err != nil {
		return err
	}
	now := time.Now().Unix()
	// 业务逻辑：软删写入 >0 的 unix 秒；前置与更新条件均用未删谓词。
	_, err := g.DB().Model(predictProIdeaTable).Ctx(ctx).
		Where("id", ideaID).Where(predictProIdeaAliveWhere).
		Data(g.Map{"deleted_at": now, "updated_at": now}).
		Update()
	return err
}

// SoftDeletePredictProIdeaByAuthor App 作者自删。
// 业务逻辑：仅当想法未删且 row.wx_id 等于调用者时写 deleted_at；禁止复用无 owner 校验的 Admin 软删。
// Args: ideaID 想法主键；wxID 会话用户。
// Returns: 不存在/已删 → ErrPredictProIdeaNotFound；非作者 → ErrPredictProIdeaForbidden。
func (s *service) SoftDeletePredictProIdeaByAuthor(ctx context.Context, ideaID, wxID int64) error {
	row, err := loadActiveIdea(ctx, ideaID)
	if err != nil {
		return err
	}
	// 业务逻辑：所有权校验失败不得改行，且不得装成功。
	if row.WxId != wxID {
		return ErrPredictProIdeaForbidden
	}
	now := time.Now().Unix()
	_, err = g.DB().Model(predictProIdeaTable).Ctx(ctx).
		Where("id", ideaID).Where(predictProIdeaAliveWhere).Where("wx_id", wxID).
		Data(g.Map{"deleted_at": now, "updated_at": now}).
		Update()
	return err
}
