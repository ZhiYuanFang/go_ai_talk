package aimodel

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

// Invoke 非流式 LLM：LoadProfile → Acquire → HTTP → release。
func Invoke(ctx context.Context, lane Lane, req ChatRequest) (ChatResponse, error) {
	profile, err := LoadProfile(ctx, lane)
	if err != nil {
		return ChatResponse{}, err
	}
	return invokeWithProfile(ctx, profile, req, false)
}

// InvokeStream 流式 LLM：持槽覆盖整段 SSE 读取。
func InvokeStream(ctx context.Context, lane Lane, req ChatRequest, cb StreamCallbacks) (StreamResult, error) {
	profile, err := LoadProfile(ctx, lane)
	if err != nil {
		return StreamResult{}, err
	}
	return invokeStreamWithProfile(ctx, profile, req, cb)
}

// InvokeWithHeldProfile 在调用方已 Acquire 后发起非流式请求（不再二次抢槽）。
func InvokeWithHeldProfile(ctx context.Context, profile Profile, req ChatRequest) (ChatResponse, error) {
	normalizeProfile(&profile)
	return invokeHTTP(ctx, profile, req, false)
}

// InvokeStreamWithHeldProfile 在调用方已 Acquire 后发起流式请求（不再二次抢槽）。
func InvokeStreamWithHeldProfile(ctx context.Context, profile Profile, req ChatRequest, cb StreamCallbacks) (StreamResult, error) {
	normalizeProfile(&profile)
	return invokeStreamHTTP(ctx, profile, req, cb)
}

func invokeWithProfile(ctx context.Context, profile Profile, req ChatRequest, stream bool) (ChatResponse, error) {
	release, err := Acquire(ctx, profile)
	if err != nil {
		return ChatResponse{}, err
	}
	defer release()
	return invokeHTTP(ctx, profile, req, stream)
}

func invokeHTTP(ctx context.Context, profile Profile, req ChatRequest, stream bool) (ChatResponse, error) {
	body, err := buildRequestBody(profile, req, stream)
	if err != nil {
		return ChatResponse{}, err
	}
	timeout := requestTimeout(profile, req)
	respBody, status, err := doHTTP(ctx, profile, body, timeout)
	if err != nil {
		return ChatResponse{}, err
	}
	if status >= 300 {
		return ChatResponse{}, fmt.Errorf("LLM HTTP %d: %s", status, truncate(string(respBody), 512))
	}
	content, err := extractChatContent(respBody)
	if err != nil {
		return ChatResponse{}, err
	}
	return ChatResponse{RawBody: respBody, Content: content}, nil
}

func invokeStreamWithProfile(ctx context.Context, profile Profile, req ChatRequest, cb StreamCallbacks) (StreamResult, error) {
	release, err := Acquire(ctx, profile)
	if err != nil {
		return StreamResult{}, err
	}
	defer release()
	return invokeStreamHTTP(ctx, profile, req, cb)
}

func invokeStreamHTTP(ctx context.Context, profile Profile, req ChatRequest, cb StreamCallbacks) (StreamResult, error) {
	body, err := buildRequestBody(profile, req, true)
	if err != nil {
		return StreamResult{}, err
	}
	timeout := requestTimeout(profile, req)
	apiKey, err := resolveAPIKey(ctx, profile.Provider)
	if err != nil {
		return StreamResult{}, err
	}
	endpoint := DefaultEndpoint(profile.Provider)
	if endpoint == "" {
		return StreamResult{}, fmt.Errorf("未知 provider: %s", profile.Provider)
	}

	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(cctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return StreamResult{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return StreamResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return StreamResult{}, fmt.Errorf("LLM HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	adapter := getProviderAdapter(profile.Provider)
	var thinkingBuf, answerBuf, contentBuf strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return StreamResult{Thinking: thinkingBuf.String(), Answer: answerBuf.String(), Content: contentBuf.String()}, err
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			continue
		}
		think, ans, content, pErr := adapter.ParseStreamDelta(payload)
		if pErr != nil {
			return StreamResult{}, pErr
		}
		if think != "" {
			thinkingBuf.WriteString(think)
			if cb.OnThinkingDelta != nil {
				if e := cb.OnThinkingDelta(think); e != nil {
					return StreamResult{Thinking: thinkingBuf.String(), Answer: answerBuf.String()}, e
				}
			}
		}
		if ans != "" {
			answerBuf.WriteString(ans)
			if cb.OnAnswerDelta != nil {
				if e := cb.OnAnswerDelta(ans); e != nil {
					return StreamResult{Thinking: thinkingBuf.String(), Answer: answerBuf.String()}, e
				}
			}
		}
		if content != "" {
			contentBuf.WriteString(content)
			if cb.OnContentDelta != nil {
				if e := cb.OnContentDelta(content); e != nil {
					return StreamResult{Content: contentBuf.String()}, e
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return StreamResult{}, err
	}
	return StreamResult{
		Thinking: thinkingBuf.String(),
		Answer:   strings.TrimSpace(answerBuf.String()),
		Content:  contentBuf.String(),
	}, nil
}

func buildRequestBody(profile Profile, req ChatRequest, stream bool) ([]byte, error) {
	payload := map[string]interface{}{
		"model":    profile.Model,
		"messages": serializeMessages(req.Messages),
		"stream":   stream,
	}
	if req.MaxTokens > 0 {
		payload["max_tokens"] = req.MaxTokens
	}
	for k, v := range req.ExtraTopLevel {
		payload[k] = v
	}
	getProviderAdapter(profile.Provider).ApplyThinkingOptions(req, payload)
	return json.Marshal(payload)
}

func serializeMessages(msgs []Message) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, map[string]interface{}{
			"role":    m.Role,
			"content": m.Content,
		})
	}
	return out
}

func requestTimeout(profile Profile, req ChatRequest) time.Duration {
	sec := req.TimeoutSec
	if sec <= 0 {
		sec = profile.TimeoutSec
	}
	if sec <= 0 {
		sec = 60
	}
	return time.Duration(sec) * time.Second
}

func doHTTP(ctx context.Context, profile Profile, body []byte, timeout time.Duration) ([]byte, int, error) {
	apiKey, err := resolveAPIKey(ctx, profile.Provider)
	if err != nil {
		return nil, 0, err
	}
	endpoint := DefaultEndpoint(profile.Provider)
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(cctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	return respBody, resp.StatusCode, err
}

func extractChatContent(body []byte) (string, error) {
	var parsed struct {
		Choices []struct {
			Message struct {
				Content          string `json:"content"`
				ReasoningContent string `json:"reasoning_content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("LLM 响应为空")
	}
	msg := parsed.Choices[0].Message
	content := strings.TrimSpace(msg.Content)
	if content == "" {
		// GLM 4.7 系列非流式响应可能将正文放在 reasoning_content。
		content = strings.TrimSpace(msg.ReasoningContent)
	}
	if content == "" {
		return "", fmt.Errorf("LLM 正文为空")
	}
	return content, nil
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n]
}
