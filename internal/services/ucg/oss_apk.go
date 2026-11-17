package ucg

import (
	"context"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// ApkObjectKeyPrefix 版本 APK 在 OSS 上的 objectKey 前缀（与社交媒资/logo 隔离）。
const ApkObjectKeyPrefix = "apk/"

// MaxApkUploadBytesDefault 单包默认上限 200MB，与 gatewayApp.apkMaxBytes 对齐。
const MaxApkUploadBytesDefault = 200 << 20

// apkContentType Android 安装包 MIME。
const apkContentType = "application/vnd.android.package-archive"

// MaxApkUploadBytes 读取 APK 上传字节上限：环境变量 UCG_APK_MAX_BYTES > 配置 ucg.apkMaxBytes > 默认 200MB。
//
// 业务：内部 upload-apk 与 gateway 转发共用此上限；须配合 ucg server.clientMaxBodySize ≥ 上限。
func MaxApkUploadBytes(ctx context.Context) int64 {
	if v := strings.TrimSpace(os.Getenv("UCG_APK_MAX_BYTES")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			return n
		}
	}
	n := g.Cfg().MustGet(ctx, "ucg.apkMaxBytes").Int64()
	if n > 0 {
		return n
	}
	return MaxApkUploadBytesDefault
}

// UploadApkObject 将 APK 流式直传至 OSS apk/ 前缀；不写 ownership / media_blob，不做缩略图。
//
// Args:
//   - body: APK 字节流（调用方可用 LimitReader 防护）
//   - size: 已知内容长度（>0）；用于 Content-Length 与上限校验
//
// Returns: objectKey（apk/...）、cdnUrl
// Side Effects: PutObject 至配置 bucket
func UploadApkObject(ctx context.Context, body io.Reader, size int64) (objectKey, cdnURL string, err error) {
	maxBytes := MaxApkUploadBytes(ctx)
	if size <= 0 || size > maxBytes {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "APK 文件大小无效或超过上限")
	}
	if body == nil {
		return "", "", gerror.NewCode(gcode.CodeInvalidParameter, "APK 内容为空")
	}
	cfg := LoadOSSConfig(ctx)
	if err = validateOSSConfig(cfg); err != nil {
		return "", "", err
	}
	objectKey = buildObjectKey(ApkObjectKeyPrefix, "apk")
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return "", "", gerror.WrapCode(gcode.CodeInternalError, err, "OSS 客户端初始化失败")
	}
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return "", "", gerror.WrapCode(gcode.CodeInternalError, err, "OSS Bucket 不可用")
	}
	// 流式 Put：避免 200MB 整文件入内存；带 Content-Length 便于 OSS 校验完整性。
	limited := io.LimitReader(body, size)
	if err = bucket.PutObject(objectKey, limited,
		oss.ContentType(apkContentType),
		oss.ContentLength(size),
	); err != nil {
		return "", "", gerror.WrapCode(gcode.CodeInternalError, err, "OSS APK 上传失败")
	}
	cdnURL = strings.TrimRight(cfg.CdnBaseURL, "/") + "/" + strings.TrimPrefix(objectKey, "/")
	return objectKey, cdnURL, nil
}

// DeleteApkObject 删除 apk/ 前缀的 OSS 对象（无 ownership 门禁；供 gateway 删版本尽力清理）。
//
// Args: objectKey 必须带 apk/ 前缀，禁止删其它前缀。
// Side Effects: DeleteObject；对象不存在时按成功处理（幂等）。
func DeleteApkObject(ctx context.Context, objectKey string) error {
	key := strings.TrimSpace(objectKey)
	key = strings.TrimPrefix(key, "/")
	if key == "" || !strings.HasPrefix(key, ApkObjectKeyPrefix) {
		return gerror.NewCode(gcode.CodeInvalidParameter, "objectKey 须为 apk/ 前缀")
	}
	// 禁止路径穿越式 key
	if strings.Contains(key, "..") {
		return gerror.NewCode(gcode.CodeInvalidParameter, "objectKey 非法")
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
	if err = bucket.DeleteObject(key); err != nil {
		// 对象已不存在时部分 SDK 仍返回 err；记录由调用方 WARN，此处仍返回错误便于观测
		return gerror.WrapCode(gcode.CodeInternalError, err, "OSS APK 删除失败")
	}
	return nil
}

// IsApkObjectKey 判断库内 download_url 是否为本能力写入的 OSS objectKey。
func IsApkObjectKey(stored string) bool {
	key := strings.TrimPrefix(strings.TrimSpace(stored), "/")
	return strings.HasPrefix(key, ApkObjectKeyPrefix)
}
