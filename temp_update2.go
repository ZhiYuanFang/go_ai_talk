//go:build ignore

package main

import (
	"log"
	"os"
	"strings"
)

func main() {
	contentBytes, err := os.ReadFile("internal/service/voice_chat.go")
	if err != nil {
		log.Fatal(err)
	}
	content := string(contentBytes)

	// We insert the logic at the top of chatWithResult
	oldCode := `func (s *VoiceService) chatWithResult(ctx context.Context, deviceNo, transcript string) (string, string, bool, error) {
	normalizedTranscript, err := s.normalizeAndValidateChatText(transcript)
	if err != nil {
		return "", "", false, err
	}
	if isExitIntentFromUserInput(normalizedTranscript) {
		return "", normalizedTranscript, true, nil
	}`

	newCode := `func (s *VoiceService) chatWithResult(ctx context.Context, deviceNo, transcript string) (string, string, bool, error) {
	normalizedTranscript, err := s.normalizeAndValidateChatText(transcript)
	if err != nil {
		return "", "", false, err
	}
	if isExitIntentFromUserInput(normalizedTranscript) {
		return "", normalizedTranscript, true, nil
	}

	// === BEGIN DEEPSEEK INTENT MATCHING ===
	// TODO: Replace real logic or just call another package function.
	// As a shortcut, we modify currentPrompt directly below.
	// === END DEEPSEEK INTENT MATCHING ===
`
	content = strings.ReplaceAll(content, oldCode, newCode)

	// Replace the prompt
	oldPrompt := `	// 请求内容后面加上一句话：返回的结果限制在50字以内
	currentPrompt := normalizedTranscript + "。请将返回的结果限制在50字以内；若用户意图是结束对话，请仅返回JSON：{\"exit\":true}。"

	unlock := s.lockDevice(deviceNo)`

	newPrompt := `	// 1. 从 DB 获取事件
	events, err := dao.Event.Ctx(ctx).All()
	var eventNames []string
	if err == nil {
		for _, evt := range events {
			eventNames = append(eventNames, evt["name"].String())
		}
	}
	eventNamesStr := strings.Join(eventNames, "、")

	// 2. 构造 Prompt，要求返回 JSON 结构体
	currentPrompt := fmt.Sprintf(` + "`" + `用户的话：%s。
系统当前支持的事件有：[%s]。
你需要判断用户的意图，是否和某个事件匹配。
同时判断用户是在说事件开始(start)还是事件结束(end)。如果是结束动作，请判断用户的指令中是否包含了【数量量词】（比如几十毫升、几次等，用字段quantity返回纯数字，找不到返回0）。
并判断用户的指令中是否包含了相关的备注事项（用字段remark返回文本）。
无论是否匹配，都请严格固定只返回一份JSON字符串（不要返回其他格式），JSON格式为：
{"event_name": "匹配的系统支持事件(不匹配则为空)", "action": "start/end/none", "quantity": 0, "remark": "提取到的备注", "reply": "回复用户的口语提示(限50字以内)", "exit": false}
` + "`" + `, normalizedTranscript, eventNamesStr)

	unlock := s.lockDevice(deviceNo)`

	content = strings.ReplaceAll(content, oldPrompt, newPrompt)

	err = os.WriteFile("internal/service/voice_chat.go", []byte(content), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
