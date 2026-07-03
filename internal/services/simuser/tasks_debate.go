package simuser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"hello/internal/services/aimodel"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

const simDebateArgumentMaxRunes = 10
const simDebateLabelMaxRunes = 5

type debatePostLLMOut struct {
	Content     string `json:"content"`
	DebateLeft  string `json:"debateLeft"`
	DebateRight string `json:"debateRight"`
}

type debateCommentLLMOut struct {
	Side     string `json:"side"`
	Argument string `json:"argument"`
}

type postSampleRow struct {
	PostId      uint64 `json:"postId"`
	AuthorWxId  int64  `json:"authorWxId"`
	Content     string `json:"content"`
	MediaType   int    `json:"mediaType"`
	DebateLeft  string `json:"debateLeft"`
	DebateRight string `json:"debateRight"`
}

// isDebateSamplePost 与 ucg isDebatePost 对齐：左右立场均非空即视为辩论帖。
func isDebateSamplePost(left, right string) bool {
	return strings.TrimSpace(left) != "" && strings.TrimSpace(right) != ""
}

func trimDebateLabelSim(label string) string {
	return strings.TrimSpace(label)
}

func validateDebateLabelSim(label string) error {
	label = trimDebateLabelSim(label)
	if label == "" {
		return fmt.Errorf("立场标签不能为空")
	}
	if utf8.RuneCountInString(label) > simDebateLabelMaxRunes {
		return fmt.Errorf("立场标签最多 %d 字", simDebateLabelMaxRunes)
	}
	return nil
}

func truncateRunes(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || s == "" {
		return s
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}

func parseLLMJSON(raw string, out interface{}) error {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)
	return json.Unmarshal([]byte(raw), out)
}

// RunPostDebateTask T7：LLM 生成辩论话题并发帖（默认 12h）。
func RunPostDebateTask(ctx context.Context, password string) {
	if !taskEnabled(ctx, "post_debate") {
		return
	}
	sess, _, err := randomSimSession(ctx, password)
	if err != nil {
		RecordTaskRun(ctx, "post_debate", false, err.Error())
		return
	}
	topic := randomTopic()
	_, user, _ := LoadRenderedPrompt(ctx, "post_debate_text", map[string]string{"topic": topic})
	resp, err := aimodel.Invoke(ctx, aimodel.LaneSimText, simChatRequest(simTempPostDebateText, 256, []aimodel.Message{{Role: "user", Content: user}}))
	if err != nil {
		RecordTaskRun(ctx, "post_debate", false, err.Error())
		return
	}
	var parsed debatePostLLMOut
	if err = parseLLMJSON(resp.Content, &parsed); err != nil {
		RecordTaskRun(ctx, "post_debate", false, "JSON 解析失败")
		return
	}
	content := strings.TrimSpace(parsed.Content)
	left := trimDebateLabelSim(parsed.DebateLeft)
	right := trimDebateLabelSim(parsed.DebateRight)
	if content == "" {
		RecordTaskRun(ctx, "post_debate", false, "话题正文为空")
		return
	}
	if err = validateDebateLabelSim(left); err != nil {
		RecordTaskRun(ctx, "post_debate", false, err.Error())
		return
	}
	if err = validateDebateLabelSim(right); err != nil {
		RecordTaskRun(ctx, "post_debate", false, err.Error())
		return
	}
	if err = appPost(ctx, sess.AccessToken, "/ucg/app/api/posts", g.Map{
		"content": content, "debateLeft": left, "debateRight": right,
		"mediaType": 0, "submit": true,
	}, nil); err != nil {
		RecordTaskRun(ctx, "post_debate", false, err.Error())
		return
	}
	RecordTaskRun(ctx, "post_debate", true, "")
	glog.Infof(ctx, "[simuser] post_debate wxId=%d left=%q right=%q", sess.WxID, left, right)
}

// RunDebateCommentTask T8：随机辩论帖投票并发表 ≤10 字论点（默认 1h）。
func RunDebateCommentTask(ctx context.Context, password string) {
	if !taskEnabled(ctx, "debate_comment") {
		return
	}
	sess, _, err := randomSimSession(ctx, password)
	if err != nil {
		RecordTaskRun(ctx, "debate_comment", false, err.Error())
		return
	}
	var sample struct {
		List []postSampleRow `json:"list"`
	}
	if err = ucgInternalPost(ctx, "/ucg/internal/api/posts/sample", g.Map{
		"mode": "random", "onlyDebate": true,
	}, &sample); err != nil || len(sample.List) == 0 {
		RecordTaskRun(ctx, "debate_comment", false, "无已发布辩论帖")
		return
	}
	post := sample.List[0]
	if !isDebateSamplePost(post.DebateLeft, post.DebateRight) {
		RecordTaskRun(ctx, "debate_comment", false, "sample 非辩论帖")
		return
	}
	if post.AuthorWxId == sess.WxID {
		RecordTaskRun(ctx, "debate_comment", false, "不可评自己的帖")
		return
	}
	_, user, _ := LoadRenderedPrompt(ctx, "debate_comment", map[string]string{
		"post_content": post.Content,
		"debate_left":  post.DebateLeft,
		"debate_right": post.DebateRight,
	})
	resp, err := aimodel.Invoke(ctx, aimodel.LaneSimText, simChatRequest(simTempDebateComment, 128, []aimodel.Message{{Role: "user", Content: user}}))
	if err != nil {
		RecordTaskRun(ctx, "debate_comment", false, err.Error())
		return
	}
	var parsed debateCommentLLMOut
	if err = parseLLMJSON(resp.Content, &parsed); err != nil {
		RecordTaskRun(ctx, "debate_comment", false, "JSON 解析失败")
		return
	}
	side := strings.ToLower(strings.TrimSpace(parsed.Side))
	if side != "left" && side != "right" {
		RecordTaskRun(ctx, "debate_comment", false, "side 无效")
		return
	}
	argument := truncateRunes(parsed.Argument, simDebateArgumentMaxRunes)
	if argument == "" {
		RecordTaskRun(ctx, "debate_comment", false, "论点为空")
		return
	}
	votePath := fmt.Sprintf("/ucg/app/api/posts/%d/vote", post.PostId)
	if err = appPost(ctx, sess.AccessToken, votePath, g.Map{"side": side}, nil); err != nil {
		RecordTaskRun(ctx, "debate_comment", false, err.Error())
		return
	}
	commentPath := fmt.Sprintf("/ucg/app/api/posts/%d/comments", post.PostId)
	if err = appPost(ctx, sess.AccessToken, commentPath, g.Map{"content": argument}, nil); err != nil {
		RecordTaskRun(ctx, "debate_comment", false, err.Error())
		return
	}
	RecordTaskRun(ctx, "debate_comment", true, "")
}
