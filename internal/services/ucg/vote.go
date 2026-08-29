package ucg

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

type voteCounts struct {
	left  uint
	right uint
}

// VotePost 对辩论帖投票；幂等、可换边。
func VotePost(ctx context.Context, wxID int64, postID uint64, side string) error {
	if wxID <= 0 || postID == 0 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "参数无效")
	}
	side = normalizeVoteSide(side)
	if side == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "side 须为 left 或 right")
	}
	post, err := loadPublishedPost(ctx, postID)
	if err != nil {
		return err
	}
	if !isDebatePost(post) {
		return gerror.NewCode(gcode.CodeInvalidParameter, "仅辩论帖可投票")
	}
	now := time.Now().Unix()
	row, err := dao.UcgPostVote.Ctx(ctx).
		Where(dao.UcgPostVote.Columns().PostId, postID).
		Where(dao.UcgPostVote.Columns().VoterWxId, wxID).
		One()
	if err != nil {
		return err
	}
	isNew := row.IsEmpty()
	oldSide := ""
	if !isNew {
		var existing entity.UcgPostVote
		if err = row.Struct(&existing); err != nil {
			return err
		}
		if existing.Side == side {
			return nil
		}
		oldSide = existing.Side
	}
	if isNew {
		_, err = dao.UcgPostVote.Ctx(ctx).Data(g.Map{
			dao.UcgPostVote.Columns().PostId:    postID,
			dao.UcgPostVote.Columns().VoterWxId: wxID,
			dao.UcgPostVote.Columns().Side:      side,
			dao.UcgPostVote.Columns().CreatedAt: now,
			dao.UcgPostVote.Columns().UpdatedAt: now,
		}).Insert()
	} else {
		var existing entity.UcgPostVote
		if err = row.Struct(&existing); err != nil {
			return err
		}
		_, err = dao.UcgPostVote.Ctx(ctx).
			Where(dao.UcgPostVote.Columns().Id, existing.Id).
			Data(g.Map{
				dao.UcgPostVote.Columns().Side:      side,
				dao.UcgPostVote.Columns().UpdatedAt: now,
			}).Update()
	}
	if err != nil {
		return err
	}
	if patchErr := patchVoteRedis(ctx, wxID, postID, oldSide, side); patchErr != nil {
		g.Log().Warningf(ctx, "[ucg-vote] patch vote redis failed wxId=%d post=%d err=%v", wxID, postID, patchErr)
		if _, bfErr := backfillVoteCountsFromMySQL(ctx, postID); bfErr != nil {
			g.Log().Warningf(ctx, "[ucg-vote] backfill after patch failed post=%d err=%v", postID, bfErr)
		}
	}
	// 作者自投：原力在 ucg 本域 +1 并写流水（不再经 device.wx）。
	if int64(post.AuthorWxId) == wxID {
		if incErr := AddDebateSelfVoteForce(ctx, wxID, int64(postID)); incErr != nil {
			g.Log().Warningf(ctx, "[ucg-vote] add force failed wxId=%d post=%d err=%v", wxID, postID, incErr)
		}
	}
	if int64(post.AuthorWxId) != wxID {
		var voteID uint64
		if isNew {
			voteRow, qErr := dao.UcgPostVote.Ctx(ctx).
				Where(dao.UcgPostVote.Columns().PostId, postID).
				Where(dao.UcgPostVote.Columns().VoterWxId, wxID).
				One()
			if qErr == nil && !voteRow.IsEmpty() {
				var v entity.UcgPostVote
				if sErr := voteRow.Struct(&v); sErr == nil {
					voteID = v.Id
				}
			}
		} else {
			var existing entity.UcgPostVote
			if sErr := row.Struct(&existing); sErr == nil {
				voteID = existing.Id
			}
		}
		sideLabel := post.DebateLeftLabel
		if side == VoteSideRight {
			sideLabel = post.DebateRightLabel
		}
		NotifyOnVote(ctx, wxID, post.AuthorWxId, postID, voteID, sideLabel)
	}
	refreshPostSnapshotFromDB(ctx, postID)
	return nil
}

func normalizeVoteSide(side string) string {
	switch side {
	case VoteSideLeft, VoteSideRight:
		return side
	default:
		return ""
	}
}

// debateVoteSideLabel 由评论快照 side 与帖 debate 标签计算展示文案（left/right 以外返回空）。
func debateVoteSideLabel(side, leftLabel, rightLabel string) string {
	switch normalizeVoteSide(side) {
	case VoteSideLeft:
		return strings.TrimSpace(leftLabel)
	case VoteSideRight:
		return strings.TrimSpace(rightLabel)
	default:
		return ""
	}
}

func normalizePostType(t string) string {
	if t == PostTypeDebate {
		return PostTypeDebate
	}
	return PostTypeMoment
}

func validateDebateLabel(label string) error {
	label = trimDebateLabel(label)
	if label == "" {
		return gerror.NewCode(gcode.CodeInvalidParameter, "立场标签不能为空")
	}
	if utf8.RuneCountInString(label) > 5 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "立场标签最多 5 字")
	}
	return nil
}

func trimDebateLabel(label string) string {
	return strings.TrimSpace(label)
}

// hasDebateLabels 左右立场标签均非空时视为辩论帖（创建推断与 vote/comment 门禁共用）。
func hasDebateLabels(left, right string) bool {
	return trimDebateLabel(left) != "" && trimDebateLabel(right) != ""
}

// isDebatePost 辩论帖判定：标签双填或 DB type=debate（兼容旧数据）。
func isDebatePost(post *entity.UcgPost) bool {
	if post == nil {
		return false
	}
	if hasDebateLabels(post.DebateLeftLabel, post.DebateRightLabel) {
		return true
	}
	return normalizePostType(post.Type) == PostTypeDebate
}

func aggregateVoteCounts(ctx context.Context, postIDs []uint64) (map[uint64]voteCounts, error) {
	out := make(map[uint64]voteCounts, len(postIDs))
	if len(postIDs) == 0 {
		return out, nil
	}
	rows, err := dao.UcgPostVote.Ctx(ctx).
		WhereIn(dao.UcgPostVote.Columns().PostId, postIDs).
		Fields(dao.UcgPostVote.Columns().PostId, dao.UcgPostVote.Columns().Side).
		All()
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		var v entity.UcgPostVote
		if err = row.Struct(&v); err != nil {
			return nil, err
		}
		c := out[v.PostId]
		switch v.Side {
		case VoteSideLeft:
			c.left++
		case VoteSideRight:
			c.right++
		}
		out[v.PostId] = c
	}
	return out, nil
}

// resolveVoteCountsForPost 读帖级 left/right 计数：Redis 优先，miss 或全 0 时与 MySQL 对齐。
func resolveVoteCountsForPost(ctx context.Context, postID uint64) voteCounts {
	if postID == 0 {
		return voteCounts{}
	}
	counts, miss, err := loadVoteCountsFromRedis(ctx, []uint64{postID})
	if err != nil {
		return voteCounts{}
	}
	reconciled, rErr := reconcileVoteCountsWithMySQL(ctx, counts, miss)
	if rErr != nil {
		return counts[postID]
	}
	return reconciled[postID]
}

func enrichPostsWithVoteData(ctx context.Context, viewerWxID int64, posts []*PostDTO) error {
	if len(posts) == 0 {
		return nil
	}
	ids := make([]uint64, 0, len(posts))
	for _, p := range posts {
		if p != nil && normalizePostType(p.Type) == PostTypeDebate {
			ids = append(ids, p.Id)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	counts, miss, err := loadVoteCountsFromRedis(ctx, ids)
	if err != nil {
		return err
	}
	counts, err = reconcileVoteCountsWithMySQL(ctx, counts, miss)
	if err != nil {
		return err
	}
	sides, err := loadMyVoteSidesFromRedis(ctx, viewerWxID, ids)
	if err != nil {
		return err
	}
	for _, p := range posts {
		if p == nil || normalizePostType(p.Type) != PostTypeDebate {
			continue
		}
		snapLeft, snapRight := p.LeftVoteCount, p.RightVoteCount
		c := counts[p.Id]
		if voteCountsTotal(c) > 0 {
			p.LeftVoteCount = c.left
			p.RightVoteCount = c.right
		} else if snapLeft > 0 || snapRight > 0 {
			// Hash 仍为 0 时保留 PostSnapshot 内嵌计数，避免 Feed 长期缺票。
			p.LeftVoteCount = snapLeft
			p.RightVoteCount = snapRight
		} else {
			p.LeftVoteCount = c.left
			p.RightVoteCount = c.right
		}
		if side, ok := sides[p.Id]; ok {
			p.MyVoteSide = side
		}
	}
	return nil
}
