package ucg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

var (
	transcodeSem     chan struct{}
	transcodeSemOnce sync.Once
	transcodeSemSize int
)

// NormalizeVideo 将输入转码为 v2 canonical mp4（H.264 Main + AAC + faststart；无音轨补静音 AAC）。
func NormalizeVideo(ctx context.Context, input []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, videoInvalidErr("视频为空")
	}
	if len(input) > MaxMediaUploadBytes {
		return nil, videoInvalidErr("视频超过大小上限")
	}
	vCfg := LoadVideoConfig(ctx)
	if err := acquireTranscode(ctx, vCfg.MaxTranscodeConcurrency); err != nil {
		return nil, gerrorWrapInternal(err, "转码排队失败")
	}
	defer releaseTranscode()

	inPath, inCleanup, err := writeTempVideoFile(input)
	if err != nil {
		return nil, gerrorWrapInternal(err, "创建输入临时文件失败")
	}
	defer inCleanup()

	outPath, err := tempOutputPath("ucg-video-out-")
	if err != nil {
		return nil, gerrorWrapInternal(err, "创建输出临时文件失败")
	}
	defer func() { _ = os.Remove(outPath) }()

	hasAudio, err := probeHasAudioStream(inPath)
	if err != nil {
		return nil, gerrorWrapInternal(err, "探测音轨失败")
	}

	timeout := time.Duration(vCfg.TranscodeTimeoutSec) * time.Second
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := runFFmpegTranscode(runCtx, inPath, outPath, hasAudio); err != nil {
		if runCtx.Err() == context.DeadlineExceeded {
			return nil, gerrorWrapInternal(err, "视频转码超时")
		}
		return nil, gerrorWrapInternal(err, "视频转码失败")
	}

	out, err := os.ReadFile(outPath)
	if err != nil {
		return nil, gerrorWrapInternal(err, "读取转码结果失败")
	}
	if err := ValidateVideoBytes(VideoTransformV2, out); err != nil {
		return nil, gerrorWrapInternal(err, "转码结果未通过 v2 验真")
	}
	return out, nil
}

func runFFmpegTranscode(ctx context.Context, inPath, outPath string, hasAudio bool) error {
	args := []string{"-y", "-i", inPath}
	if !hasAudio {
		args = append(args, "-f", "lavfi", "-i", "anullsrc=channel_layout=stereo:sample_rate=44100")
	}
	args = append(args,
		"-c:v", "libx264", "-profile:v", "main", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-b:a", "128k",
	)
	if !hasAudio {
		args = append(args, "-shortest")
	}
	args = append(args, "-movflags", "+faststart", outPath)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, truncateFFmpegLog(out))
	}
	return nil
}

func truncateFFmpegLog(b []byte) string {
	const max = 512
	s := string(b)
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func acquireTranscode(ctx context.Context, max int) error {
	transcodeSemOnce.Do(func() {
		n := max
		if n <= 0 {
			n = 2
		}
		transcodeSemSize = n
		transcodeSem = make(chan struct{}, n)
	})
	if max > 0 && max != transcodeSemSize {
		// 配置变更不重建 channel；首次加载后固定容量。
	}
	select {
	case transcodeSem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func releaseTranscode() {
	if transcodeSem != nil {
		<-transcodeSem
	}
}

func gerrorWrapInternal(err error, msg string) error {
	if err == nil {
		return nil
	}
	return gerror.WrapCode(gcode.CodeInternalError, err, msg)
}
