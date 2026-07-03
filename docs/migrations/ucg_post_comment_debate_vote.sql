-- 辩论评论发帖时投票立场快照（读路径不 join ucg_post_vote）
ALTER TABLE ucg_post_comment
  ADD COLUMN debate_vote_side VARCHAR(8) NULL COMMENT 'left/right 辩论评论发帖时投票立场快照';

-- 历史已发布评论 best-effort 回填（依赖当时投票记录，换边后旧评论仍保留原立场）
UPDATE ucg_post_comment c
  INNER JOIN ucg_post p ON p.id = c.post_id AND p.type = 'debate'
  INNER JOIN ucg_post_vote v ON v.post_id = c.post_id AND v.voter_wx_id = c.author_wx_id
SET c.debate_vote_side = v.side
WHERE c.status = 2 AND (c.debate_vote_side IS NULL OR c.debate_vote_side = '');
