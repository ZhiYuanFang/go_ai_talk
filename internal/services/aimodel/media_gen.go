package aimodel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	zhipuImageGenerationsURL = "https://open.bigmodel.cn/api/paas/v4/images/generations"
	zhipuVideoGenerationsURL = "https://open.bigmodel.cn/api/paas/v4/videos/generations"
)

// ImageGenerationResult CogView 生图结果。
type ImageGenerationResult struct {
	URL string
}

// VideoGenerationStatus 视频生成任务状态。
type VideoGenerationStatus string

const (
	VideoStatusProcessing VideoGenerationStatus = "processing"
	VideoStatusSuccess    VideoGenerationStatus = "success"
	VideoStatusFailed     VideoGenerationStatus = "failed"
)

// VideoPollResult 视频轮询结果。
type VideoPollResult struct {
	Status   VideoGenerationStatus
	VideoURL string
}

// GenerateImage 经 lane 调用智谱 CogView 生图（Acquire 覆盖上游 POST）。
func GenerateImage(ctx context.Context, lane Lane, prompt string) (ImageGenerationResult, error) {
	profile, err := LoadProfile(ctx, lane)
	if err != nil {
		return ImageGenerationResult{}, err
	}
	release, err := Acquire(ctx, profile)
	if err != nil {
		return ImageGenerationResult{}, err
	}
	defer release()
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return ImageGenerationResult{}, fmt.Errorf("aimodel: 生图 prompt 为空")
	}
	body, _ := json.Marshal(map[string]interface{}{
		"model":  profile.Model,
		"prompt": prompt,
		// 智谱 CogView 默认带水印；关闭需账户已签去水印免责声明。
		"watermark_enabled": false,
	})
	respBody, status, err := doZhipuMediaHTTP(ctx, profile, http.MethodPost, zhipuImageGenerationsURL, body)
	if err != nil {
		return ImageGenerationResult{}, err
	}
	if status >= 300 {
		return ImageGenerationResult{}, fmt.Errorf("生图 HTTP %d: %s", status, truncate(string(respBody), 512))
	}
	url, err := extractImageURL(respBody)
	if err != nil {
		return ImageGenerationResult{}, err
	}
	return ImageGenerationResult{URL: url}, nil
}

// SubmitVideoGeneration 提交 CogVideoX 异步视频生成任务。
func SubmitVideoGeneration(ctx context.Context, lane Lane, prompt string) (taskID string, err error) {
	profile, err := LoadProfile(ctx, lane)
	if err != nil {
		return "", err
	}
	release, err := Acquire(ctx, profile)
	if err != nil {
		return "", err
	}
	defer release()
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return "", fmt.Errorf("aimodel: 生视频 prompt 为空")
	}
	body, _ := json.Marshal(map[string]interface{}{
		"model":  profile.Model,
		"prompt": prompt,
	})
	respBody, status, err := doZhipuMediaHTTP(ctx, profile, http.MethodPost, zhipuVideoGenerationsURL, body)
	if err != nil {
		return "", err
	}
	if status >= 300 {
		return "", fmt.Errorf("生视频 HTTP %d: %s", status, truncate(string(respBody), 512))
	}
	return extractVideoTaskID(respBody)
}

// PollVideoGeneration 轮询视频任务（不占 inflight 槽）。
func PollVideoGeneration(ctx context.Context, taskID string) (VideoPollResult, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return VideoPollResult{}, fmt.Errorf("aimodel: taskID 为空")
	}
	apiKey, err := resolveAPIKey(ctx, ProviderZhipu)
	if err != nil {
		return VideoPollResult{}, err
	}
	url := zhipuVideoGenerationsURL + "/" + taskID
	cctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(cctx, http.MethodGet, url, nil)
	if err != nil {
		return VideoPollResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return VideoPollResult{}, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return VideoPollResult{}, err
	}
	if resp.StatusCode >= 300 {
		return VideoPollResult{}, fmt.Errorf("视频轮询 HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 512))
	}
	return parseVideoPollBody(respBody)
}

func doZhipuMediaHTTP(ctx context.Context, profile Profile, method, url string, body []byte) ([]byte, int, error) {
	apiKey, err := resolveAPIKey(ctx, profile.Provider)
	if err != nil {
		return nil, 0, err
	}
	timeout := requestTimeout(profile, ChatRequest{})
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(cctx, method, url, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	return respBody, resp.StatusCode, err
}

func extractImageURL(body []byte) (string, error) {
	var parsed struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Data) == 0 || strings.TrimSpace(parsed.Data[0].URL) == "" {
		return "", fmt.Errorf("生图响应无 url")
	}
	return strings.TrimSpace(parsed.Data[0].URL), nil
}

func extractVideoTaskID(body []byte) (string, error) {
	var parsed struct {
		ID     string `json:"id"`
		TaskID string `json:"task_id"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	id := strings.TrimSpace(parsed.ID)
	if id == "" {
		id = strings.TrimSpace(parsed.TaskID)
	}
	if id == "" {
		return "", fmt.Errorf("生视频响应无 task id")
	}
	return id, nil
}

func parseVideoPollBody(body []byte) (VideoPollResult, error) {
	var parsed struct {
		TaskStatus string `json:"task_status"`
		Status     string `json:"status"`
		VideoURL   string `json:"video_url"`
		Output     struct {
			VideoURL string `json:"video_url"`
		} `json:"output"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return VideoPollResult{}, err
	}
	st := strings.ToLower(strings.TrimSpace(parsed.TaskStatus))
	if st == "" {
		st = strings.ToLower(strings.TrimSpace(parsed.Status))
	}
	videoURL := strings.TrimSpace(parsed.VideoURL)
	if videoURL == "" {
		videoURL = strings.TrimSpace(parsed.Output.VideoURL)
	}
	switch st {
	case "succeed", "success", "completed", "done":
		if videoURL == "" {
			return VideoPollResult{Status: VideoStatusProcessing}, nil
		}
		return VideoPollResult{Status: VideoStatusSuccess, VideoURL: videoURL}, nil
	case "failed", "fail", "error":
		return VideoPollResult{Status: VideoStatusFailed}, nil
	default:
		return VideoPollResult{Status: VideoStatusProcessing}, nil
	}
}
