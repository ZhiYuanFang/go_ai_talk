package voice

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type clinicStreamCallbacks struct {
	OnThinkingDelta func(delta string) error
	OnAnswerDelta   func(delta string) error
}

type clinicLLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// streamClinicLLM 调用 DeepSeek 流式接口，区分 reasoning/content 映射 thinking_delta / answer_delta。
func (s *ClinicService) streamClinicLLM(ctx context.Context, summaryJSON, question string, prior []map[string]string, cb clinicStreamCallbacks) (thinking, answer string, err error) {
	cfg := s.cfg
	if strings.TrimSpace(cfg.Endpoint) == "" || strings.TrimSpace(cfg.APIKey) == "" {
		return "", "", fmt.Errorf("DeepSeek 未配置")
	}
	system := strings.TrimSpace(cfg.SystemPrompt)
	if system == "" {
		system = "你是「胖宝」AI 育儿助手。"
	}
	system += "\n\n近7天喂养事件聚合摘要（JSON）：\n" + summaryJSON

	messages := make([]clinicLLMMessage, 0, 2+len(prior)+1)
	messages = append(messages, clinicLLMMessage{Role: "system", Content: system})
	for _, m := range prior {
		role := strings.TrimSpace(m["role"])
		content := strings.TrimSpace(m["content"])
		if role == "" || content == "" {
			continue
		}
		messages = append(messages, clinicLLMMessage{Role: role, Content: content})
	}
	messages = append(messages, clinicLLMMessage{Role: "user", Content: strings.TrimSpace(question)})

	body := map[string]interface{}{
		"model":    cfg.Model,
		"messages": messages,
		"stream":   true,
	}
	if cfg.ThinkingEnabled {
		body["reasoning_effort"] = cfg.ReasoningEffort
		body["extra_body"] = map[string]interface{}{
			"thinking": map[string]interface{}{"type": "enabled"},
		}
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", "", err
	}

	timeout := cfg.LLMTimeoutSeconds
	if timeout <= 0 {
		timeout = 120
	}
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, cfg.Endpoint, bytes.NewReader(raw))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", "", fmt.Errorf("DeepSeek HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	var thinkingBuf, answerBuf strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			continue
		}
		thinkDelta, ansDelta, pErr := extractClinicStreamDeltas(payload)
		if pErr != nil {
			return "", "", pErr
		}
		if thinkDelta != "" {
			thinkingBuf.WriteString(thinkDelta)
			if cb.OnThinkingDelta != nil {
				if err := cb.OnThinkingDelta(thinkDelta); err != nil {
					return thinkingBuf.String(), answerBuf.String(), err
				}
			}
		}
		if ansDelta != "" {
			answerBuf.WriteString(ansDelta)
			if cb.OnAnswerDelta != nil {
				if err := cb.OnAnswerDelta(ansDelta); err != nil {
					return thinkingBuf.String(), answerBuf.String(), err
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return thinkingBuf.String(), answerBuf.String(), err
	}
	return thinkingBuf.String(), strings.TrimSpace(answerBuf.String()), nil
}

func extractClinicStreamDeltas(data string) (thinking, answer string, err error) {
	var obj map[string]interface{}
	if err = json.Unmarshal([]byte(data), &obj); err != nil {
		return "", "", err
	}
	choices, ok := obj["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", "", nil
	}
	first, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", "", nil
	}
	delta, ok := first["delta"].(map[string]interface{})
	if !ok {
		return "", "", nil
	}
	if v, ok := delta["reasoning_content"].(string); ok && v != "" {
		thinking = v
	}
	if thinking == "" {
		if v, ok := delta["thinking"].(string); ok && v != "" {
			thinking = v
		}
	}
	if v, ok := delta["content"].(string); ok && v != "" {
		answer = v
	}
	return thinking, answer, nil
}
