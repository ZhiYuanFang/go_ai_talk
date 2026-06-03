package ucg

import (
	"context"
	"time"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// auditPost 对单条 pending 帖子执行 Green 审核并更新状态。
func auditPost(ctx context.Context, post entity.UcgPost) error {
	moderator := EffectiveGreen()
	cfg := LoadOSSConfig(ctx)

	if verdict, err := moderator.ModerateText(ctx, "comment_detection", post.Content); err != nil {
		return err
	} else if !verdict.Pass {
		return rejectPost(ctx, post.Id, verdict.Reason)
	}

	media, err := loadPostMedia(ctx, post.Id)
	if err != nil {
		return err
	}
	for _, m := range media {
		url := m.CdnUrl
		if url == "" && m.ObjectKey != "" {
			url = cfg.CdnBaseURL + "/" + m.ObjectKey
		}
		var verdict AuditVerdict
		switch {
		case m.MediaKind == 2:
			verdict, err = moderator.ModerateVideoURL(ctx, url)
		default:
			verdict, err = moderator.ModerateImageURL(ctx, url)
		}
		if err != nil {
			return err
		}
		if !verdict.Pass {
			return rejectPost(ctx, post.Id, verdict.Reason)
		}
	}
	return publishPost(ctx, post.Id)
}

func publishPost(ctx context.Context, postID uint64) error {
	now := time.Now().Unix()
	_, err := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).Data(g.Map{
		dao.UcgPost.Columns().Status:       PostStatusPublished,
		dao.UcgPost.Columns().PublishedAt: now,
		dao.UcgPost.Columns().UpdatedAt:    now,
		dao.UcgPost.Columns().RejectReason: "",
	}).Update()
	return err
}

func rejectPost(ctx context.Context, postID uint64, reason string) error {
	if reason == "" {
		reason = rejectReasonDefault
	}
	_, err := dao.UcgPost.Ctx(ctx).Where(dao.UcgPost.Columns().Id, postID).Data(g.Map{
		dao.UcgPost.Columns().Status:       PostStatusRejected,
		dao.UcgPost.Columns().RejectReason: reason,
		dao.UcgPost.Columns().UpdatedAt:    time.Now().Unix(),
	}).Update()
	return err
}

func auditProfilePatch(ctx context.Context, wxID int64) error {
	patch, ok, err := LoadProfilePending(ctx, wxID)
	if err != nil || !ok {
		return err
	}
	moderator := EffectiveGreen()
	cfg := LoadOSSConfig(ctx)

	if patch.Nickname != "" {
		if verdict, mErr := moderator.ModerateText(ctx, "nickname_detection", patch.Nickname); mErr != nil {
			return mErr
		} else if !verdict.Pass {
			return failProfileAudit(ctx, wxID, verdict.Reason)
		}
	}
	if patch.Bio != "" {
		if verdict, mErr := moderator.ModerateText(ctx, "comment_detection", patch.Bio); mErr != nil {
			return mErr
		} else if !verdict.Pass {
			return failProfileAudit(ctx, wxID, verdict.Reason)
		}
	}
	if patch.AvatarKey != "" {
		url := cfg.CdnBaseURL + "/" + patch.AvatarKey
		if verdict, mErr := moderator.ModerateImageURL(ctx, url); mErr != nil {
			return mErr
		} else if !verdict.Pass {
			return failProfileAudit(ctx, wxID, verdict.Reason)
		}
	}
	if err = applyProfilePending(ctx, patch); err != nil {
		return err
	}
	return clearProfilePending(ctx, wxID)
}

func failProfileAudit(ctx context.Context, wxID int64, reason string) error {
	if reason == "" {
		reason = rejectReasonDefault
	}
	_ = setProfileRejectReason(ctx, wxID, reason)
	return clearProfilePending(ctx, wxID)
}

func listPendingAuditPosts(ctx context.Context, limit int) ([]entity.UcgPost, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := dao.UcgPost.Ctx(ctx).
		Where(dao.UcgPost.Columns().Status, PostStatusPendingAudit).
		OrderAsc(dao.UcgPost.Columns().UpdatedAt).
		Limit(limit).
		All()
	if err != nil {
		return nil, err
	}
	out := make([]entity.UcgPost, 0, len(rows))
	for _, row := range rows {
		var post entity.UcgPost
		if err = row.Struct(&post); err != nil {
			return nil, err
		}
		out = append(out, post)
	}
	return out, nil
}
