package ucg

import (
	"context"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// DeleteOwnedMedia deletes OSS objects owned by wxID unless referenced by ucg_post_media.
func DeleteOwnedMedia(ctx context.Context, wxID int64, objectKeys []string) (deleted, skipped []string, err error) {
	if wxID <= 0 {
		return nil, nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	cfg := LoadOSSConfig(ctx)
	if err = validateOSSConfig(cfg); err != nil {
		return nil, nil, err
	}
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, nil, gerror.WrapCode(gcode.CodeInternalError, err, "OSS 客户端初始化失败")
	}
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, nil, gerror.WrapCode(gcode.CodeInternalError, err, "OSS Bucket 不可用")
	}

	seen := make(map[string]struct{}, len(objectKeys))
	for _, raw := range objectKeys {
		key := strings.TrimSpace(raw)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		owned, ownErr := mediaUploadOwnedBy(ctx, wxID, key)
		if ownErr != nil {
			return deleted, skipped, ownErr
		}
		if !owned {
			skipped = append(skipped, key)
			continue
		}

		referenced, refErr := objectKeyReferencedByPost(ctx, key)
		if refErr != nil {
			return deleted, skipped, refErr
		}
		if referenced {
			skipped = append(skipped, key)
			continue
		}

		if delErr := bucket.DeleteObject(key); delErr != nil {
			return deleted, skipped, gerror.WrapCode(gcode.CodeInternalError, delErr, "OSS 删除失败")
		}
		if rowErr := deleteMediaUploadRow(ctx, wxID, key); rowErr != nil {
			return deleted, skipped, rowErr
		}
		deleted = append(deleted, key)
	}
	return deleted, skipped, nil
}
