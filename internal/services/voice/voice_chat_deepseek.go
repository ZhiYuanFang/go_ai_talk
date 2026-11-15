package voice

import (
	"context"
	"errors"
	"strings"

	"github.com/gogf/gf/v2/os/glog"
)

// callDeepSeekHistoryReply 历史问答：独立 AnalyzeIntent，省略 model，由 Python 定死 LLM。
// 业务说明：不走额度/VIP/Acquire；不得附带喂养澄清 conversation_id。
func (s *VoiceService) callDeepSeekHistoryReply(ctx context.Context, deviceNo, transcript string, hours int) (string, error) {
	_ = hours
	pythonClient := PythonAIClientFromCfg()
	pythonResp, pythonErr := pythonClient.AnalyzeIntent(ctx, &AnalyzeIntentRequest{
		Text:     transcript,
		DeviceNo: deviceNo,
	})
	if pythonErr == nil && pythonResp != nil && pythonResp.Content != "" {
		reply := strings.TrimSpace(pythonResp.Content)
		if reply != "" {
			return reply, nil
		}
	}
	if pythonErr != nil {
		glog.Warningf(ctx, "[Python AI] 历史问答调用失败。deviceNo=%s err=%v", deviceNo, pythonErr)
		return "", pythonErr
	}
	return "", errors.New("历史问答无有效内容")
}

// extractChatReply 从 DeepSeek 风格返回中提取回复文本和退出标记。
func extractChatReply(body []byte) (string, bool, error) {
	_, reply, exit, err := extractChatReplyRaw(body)
	return reply, exit, err
}

// extractChatReplyRaw 返回原始 content 与标准化 reply（避免上层丢失 JSON 结构）。
func extractChatReplyRaw(body []byte) (rawContent string, reply string, exit bool, err error) {
	rawContent, exit, err = extractChatRawContent(body)
	if err != nil || exit {
		return rawContent, "", exit, err
	}
	reply, modelExit := normalizeReplyAndDetectExit(rawContent)
	return rawContent, reply, modelExit, nil
}
