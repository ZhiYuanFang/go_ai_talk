package device

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"hello/internal/dao"
	"hello/internal/model/entity"
	"hello/internal/services/contracts"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

const (
	feedbackStatusPending   = 0
	feedbackStatusReplied   = 1
	feedbackMaxQuestionLen  = 2000
	feedbackDefaultPageSize = 20
	feedbackMaxPageSize     = 100
)

var (
	ErrFeedbackNotFound       = errors.New("反馈不存在")
	ErrFeedbackAlreadyReplied = errors.New("该反馈已回复，不可再次回复")
)

func validateFeedbackText(text string, emptyMsg string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errors.New(emptyMsg)
	}
	if utf8.RuneCountInString(text) > feedbackMaxQuestionLen {
		return "", errors.New("内容不能超过 2000 字")
	}
	return text, nil
}

func normalizeFeedbackPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = feedbackDefaultPageSize
	}
	if pageSize > feedbackMaxPageSize {
		pageSize = feedbackMaxPageSize
	}
	return page, pageSize
}

// ListFeedbackByWxID 返回指定用户的全部反馈，created_at 倒序。
func (s *service) ListFeedbackByWxID(ctx context.Context, wxID int64) ([]entity.Feedback, error) {
	if wxID <= 0 {
		return nil, errors.New("wx_id 无效")
	}
	rows := make([]entity.Feedback, 0)
	err := dao.Feedback.Ctx(ctx).
		Where(dao.Feedback.Columns().WxId, wxID).
		OrderDesc(dao.Feedback.Columns().CreatedAt).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// SubmitFeedback 插入待回复反馈。
func (s *service) SubmitFeedback(ctx context.Context, wxID int64, question string) (entity.Feedback, error) {
	if wxID <= 0 {
		return entity.Feedback{}, errors.New("wx_id 无效")
	}
	q, err := validateFeedbackText(question, "问题不能为空")
	if err != nil {
		return entity.Feedback{}, err
	}
	now := gtime.Now()
	id, err := dao.Feedback.Ctx(ctx).Data(g.Map{
		dao.Feedback.Columns().WxId:     wxID,
		dao.Feedback.Columns().Question: q,
		dao.Feedback.Columns().Status:   feedbackStatusPending,
		dao.Feedback.Columns().CreatedAt: now,
		dao.Feedback.Columns().UpdatedAt: now,
	}).InsertAndGetId()
	if err != nil {
		return entity.Feedback{}, err
	}
	var row entity.Feedback
	if err := dao.Feedback.Ctx(ctx).Where(dao.Feedback.Columns().Id, id).Scan(&row); err != nil {
		return entity.Feedback{}, err
	}
	return row, nil
}

// ListFeedbackPage Admin 分页列表；unrepliedOnly 为 true 时仅 status=0。
func (s *service) ListFeedbackPage(ctx context.Context, page, pageSize int, unrepliedOnly bool) (contracts.FeedbackPageResult, error) {
	page, pageSize = normalizeFeedbackPage(page, pageSize)
	m := dao.Feedback.Ctx(ctx)
	if unrepliedOnly {
		m = m.Where(dao.Feedback.Columns().Status, feedbackStatusPending)
	}
	total, err := m.Count()
	if err != nil {
		return contracts.FeedbackPageResult{}, err
	}
	if total == 0 {
		return contracts.FeedbackPageResult{
			List:     []entity.Feedback{},
			Total:    0,
			Page:     page,
			PageSize: pageSize,
		}, nil
	}
	rows := make([]entity.Feedback, 0)
	err = m.OrderDesc(dao.Feedback.Columns().CreatedAt).
		Page(page, pageSize).
		Scan(&rows)
	if err != nil {
		return contracts.FeedbackPageResult{}, err
	}
	return contracts.FeedbackPageResult{
		List:     rows,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ReplyFeedback 首次官方回复；已回复记录拒绝二次回复。
func (s *service) ReplyFeedback(ctx context.Context, id int64, officialReply string) error {
	if id <= 0 {
		return errors.New("id 无效")
	}
	reply, err := validateFeedbackText(officialReply, "回复不能为空")
	if err != nil {
		return err
	}
	var row entity.Feedback
	if err := dao.Feedback.Ctx(ctx).Where(dao.Feedback.Columns().Id, id).Scan(&row); err != nil {
		return err
	}
	if row.Id == 0 {
		return ErrFeedbackNotFound
	}
	if row.Status == feedbackStatusReplied || strings.TrimSpace(row.OfficialReply) != "" {
		return ErrFeedbackAlreadyReplied
	}
	now := gtime.Now()
	result, err := dao.Feedback.Ctx(ctx).
		Where(dao.Feedback.Columns().Id, id).
		Where(dao.Feedback.Columns().Status, feedbackStatusPending).
		Data(g.Map{
			dao.Feedback.Columns().OfficialReply: reply,
			dao.Feedback.Columns().Status:        feedbackStatusReplied,
			dao.Feedback.Columns().UpdatedAt:     now,
			dao.Feedback.Columns().RepliedAt:     now,
		}).Update()
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrFeedbackAlreadyReplied
	}
	return nil
}
