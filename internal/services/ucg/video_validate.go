package ucg

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// videoProbeResult ffprobe JSON 解析子集。
type videoProbeResult struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Profile   string `json:"profile"`
		PixFmt    string `json:"pix_fmt"`
	} `json:"streams"`
	Format struct {
		FormatName string `json:"format_name"`
		Size       string `json:"size"`
	} `json:"format"`
}

// ValidateVideoBytes 对内存字节按 transformVersion 验真（v1 宽规 / v2 严规）。
// Web 代理上传：v1 通过则直传；v1 失败时由 ProbeVideoDecodable 决定是否进入服务端转码（Phase 2 B 分支）。
func ValidateVideoBytes(version string, data []byte) error {
	if len(data) == 0 {
		return videoInvalidErr("视频为空")
	}
	if len(data) > MaxMediaUploadBytes {
		return videoInvalidErr("视频超过大小上限")
	}
	path, cleanup, err := writeTempVideoFile(data)
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "创建临时文件失败")
	}
	defer cleanup()
	probe, err := runFFprobe(path)
	if err != nil {
		return videoInvalidErr("无法解析视频")
	}
	if err := checkProbeAgainstVersion(version, probe, data); err != nil {
		return err
	}
	return nil
}

// ProbeVideoDecodable 探测上传字节是否可被 ffprobe 解析且含视频轨。
// 仅用于 Web Phase 2：v1 验真失败后判断是否进入 NormalizeVideo 转码兜底；不替代 v1/v2 合规验真。
func ProbeVideoDecodable(data []byte) error {
	if len(data) == 0 {
		return videoInvalidErr("视频为空")
	}
	if len(data) > MaxMediaUploadBytes {
		return videoInvalidErr("视频超过大小上限")
	}
	path, cleanup, err := writeTempVideoFile(data)
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "创建临时文件失败")
	}
	defer cleanup()
	probe, err := runFFprobe(path)
	if err != nil {
		return videoInvalidErr("无法解析视频")
	}
	for _, s := range probe.Streams {
		if strings.EqualFold(strings.TrimSpace(s.CodecType), "video") {
			return nil
		}
	}
	return videoInvalidErr("无法解析视频")
}

// ValidateVideoOnOSS 下载 OSS 对象（上限 MaxMediaUploadBytes）并按版本验真。
func ValidateVideoOnOSS(ctx context.Context, version, objectKey string) error {
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" {
		return videoInvalidErr("objectKey 无效")
	}
	cfg := LoadOSSConfig(ctx)
	if err := validateOSSConfig(cfg); err != nil {
		return err
	}
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "OSS 客户端初始化失败")
	}
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "OSS Bucket 不可用")
	}
	path, cleanup, err := downloadOSSObjectToTemp(ctx, bucket, objectKey)
	if err != nil {
		return err
	}
	defer cleanup()
	data, err := os.ReadFile(path)
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "读取 OSS 对象失败")
	}
	probe, err := runFFprobe(path)
	if err != nil {
		return videoInvalidErr("无法解析视频")
	}
	return checkProbeAgainstVersion(version, probe, data)
}

func videoInvalidErr(detail string) error {
	msg := "视频格式不合规"
	if strings.TrimSpace(detail) != "" {
		msg = msg + "：" + strings.TrimSpace(detail)
	}
	return gerror.NewCode(gcode.CodeInvalidParameter, msg)
}

func writeTempVideoFile(data []byte) (path string, cleanup func(), err error) {
	f, err := os.CreateTemp("", "ucg-video-in-*.mp4")
	if err != nil {
		return "", nil, err
	}
	path = f.Name()
	cleanup = func() { _ = os.Remove(path) }
	if _, err = f.Write(data); err != nil {
		cleanup()
		return "", nil, err
	}
	if err = f.Close(); err != nil {
		cleanup()
		return "", nil, err
	}
	return path, cleanup, nil
}

func downloadOSSObjectToTemp(ctx context.Context, bucket *oss.Bucket, objectKey string) (path string, cleanup func(), err error) {
	props, err := bucket.GetObjectMeta(objectKey, oss.WithContext(ctx))
	if err != nil {
		return "", nil, gerror.WrapCode(gcode.CodeInternalError, err, "OSS HEAD 失败")
	}
	size := parseOSSContentLength(props)
	if size > MaxMediaUploadBytes {
		return "", nil, videoInvalidErr("视频超过大小上限")
	}
	f, err := os.CreateTemp("", "ucg-video-oss-*.mp4")
	if err != nil {
		return "", nil, err
	}
	path = f.Name()
	cleanup = func() { _ = os.Remove(path) }
	_ = f.Close()
	if err = bucket.GetObjectToFile(objectKey, path, oss.WithContext(ctx)); err != nil {
		cleanup()
		return "", nil, gerror.WrapCode(gcode.CodeInternalError, err, "OSS 下载失败")
	}
	info, err := os.Stat(path)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	if info.Size() > MaxMediaUploadBytes {
		cleanup()
		return "", nil, videoInvalidErr("视频超过大小上限")
	}
	return path, cleanup, nil
}

func parseOSSContentLength(props http.Header) int64 {
	if props == nil {
		return 0
	}
	if v := props.Get("Content-Length"); v != "" {
		var n int64
		fmt.Sscanf(v, "%d", &n)
		return n
	}
	return 0
}

func runFFprobe(path string) (*videoProbeResult, error) {
	out, err := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_format", "-show_streams", path).Output()
	if err != nil {
		return nil, err
	}
	var probe videoProbeResult
	if err := json.Unmarshal(out, &probe); err != nil {
		return nil, err
	}
	return &probe, nil
}

func checkProbeAgainstVersion(version string, probe *videoProbeResult, raw []byte) error {
	if probe == nil {
		return videoInvalidErr("探测结果为空")
	}
	formatName := strings.ToLower(probe.Format.FormatName)
	if !strings.Contains(formatName, "mp4") {
		return videoInvalidErr("容器须为 mp4")
	}

	var videoOK bool
	var hasAAC bool
	for _, s := range probe.Streams {
		switch strings.ToLower(strings.TrimSpace(s.CodecType)) {
		case "video":
			if err := checkVideoStream(version, s); err != nil {
				return err
			}
			videoOK = true
		case "audio":
			if strings.EqualFold(strings.TrimSpace(s.CodecName), "aac") {
				hasAAC = true
			} else if strings.TrimSpace(s.CodecName) != "" {
				return videoInvalidErr("音频须为 AAC")
			}
		}
	}
	if !videoOK {
		return videoInvalidErr("缺少 h264 视频轨")
	}
	if !hasAAC {
		return videoInvalidErr("必须有 AAC 音轨")
	}
	if version == VideoTransformV2 && !mp4FastStart(raw) {
		return videoInvalidErr("须为 faststart mp4")
	}
	return nil
}

func checkVideoStream(version string, s struct {
	CodecType string `json:"codec_type"`
	CodecName string `json:"codec_name"`
	Profile   string `json:"profile"`
	PixFmt    string `json:"pix_fmt"`
}) error {
	if !strings.EqualFold(strings.TrimSpace(s.CodecName), "h264") {
		return videoInvalidErr("视频须为 H.264")
	}
	pix := strings.ToLower(strings.TrimSpace(s.PixFmt))
	if pix != "" && pix != "yuv420p" {
		return videoInvalidErr("像素格式须为 yuv420p")
	}
	profile := normalizeH264Profile(s.Profile)
	if version == VideoTransformV2 {
		if profile != "main" {
			return videoInvalidErr("视频 profile 须为 Main")
		}
		return nil
	}
	if profile != "main" && profile != "baseline" {
		return videoInvalidErr("视频 profile 须为 Main 或 Baseline")
	}
	return nil
}

func normalizeH264Profile(profile string) string {
	p := strings.ToLower(strings.TrimSpace(profile))
	p = strings.ReplaceAll(p, " ", "")
	if strings.Contains(p, "constrainedbaseline") || p == "baseline" {
		return "baseline"
	}
	if strings.Contains(p, "main") {
		return "main"
	}
	return p
}

// mp4FastStart 判断顶层 moov 是否在 mdat 之前（渐进播放）。
func mp4FastStart(data []byte) bool {
	if len(data) < 16 {
		return false
	}
	var moov, mdat int64 = -1, -1
	off := 0
	for off+8 <= len(data) {
		size := int(binary.BigEndian.Uint32(data[off : off+4]))
		typ := string(data[off+4 : off+8])
		if size < 8 {
			break
		}
		if size == 1 && off+16 <= len(data) {
			// extended size
			size = int(binary.BigEndian.Uint64(data[off+8 : off+16]))
		}
		boxStart := int64(off)
		next := off + size
		if next <= off || next > len(data) {
			break
		}
		switch typ {
		case "moov":
			moov = boxStart
		case "mdat":
			mdat = boxStart
		}
		off = next
	}
	if moov < 0 {
		return false
	}
	if mdat < 0 {
		return true
	}
	return moov < mdat
}

// probeHasAudioStream 探测输入是否含音轨（用于转码是否补静音 AAC）。
func probeHasAudioStream(path string) (bool, error) {
	probe, err := runFFprobe(path)
	if err != nil {
		return false, err
	}
	for _, s := range probe.Streams {
		if strings.EqualFold(strings.TrimSpace(s.CodecType), "audio") {
			return true, nil
		}
	}
	return false, nil
}

// readAllLimited 读取 reader 并限制最大字节。
func readAllLimited(r io.Reader, max int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > max {
		return nil, fmt.Errorf("超过大小上限")
	}
	return data, nil
}

// tempOutputPath 生成临时输出 mp4 路径。
func tempOutputPath(prefix string) (string, error) {
	f, err := os.CreateTemp("", prefix+"*.mp4")
	if err != nil {
		return "", err
	}
	path := f.Name()
	_ = f.Close()
	return path, nil
}
