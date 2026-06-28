package ucg

import (
	"context"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"hello/internal/dao"
	"hello/internal/model/entity"
)

// DeleteOwnedMedia deletes OSS objects owned by wxID unless referenced by ucg_post_media.
// When a blob row exists, ref_count gates OSS deletion; legacy keys without blob rows delete OSS directly.
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

		if rowErr := deleteMediaUploadRow(ctx, wxID, key); rowErr != nil {
			return deleted, skipped, rowErr
		}
		deleted = append(deleted, key)

		blob, blobErr := findMediaBlobByObjectKey(ctx, key)
		if blobErr != nil {
			return deleted, skipped, blobErr
		}
		mediaKind := 0
		if blob != nil {
			mediaKind = blob.MediaKind
		}
		if blob == nil {
			if delErr := bucket.DeleteObject(key); delErr != nil {
				return deleted, skipped, gerror.WrapCode(gcode.CodeInternalError, delErr, "OSS 删除失败")
			}
			if thumbErr := deletePairedThumbObject(bucket, key, mediaKind); thumbErr != nil {
				return deleted, skipped, thumbErr
			}
			continue
		}

		newRefCount := blob.RefCount - 1
		if newRefCount <= 0 {
			if delErr := bucket.DeleteObject(key); delErr != nil {
				return deleted, skipped, gerror.WrapCode(gcode.CodeInternalError, delErr, "OSS 删除失败")
			}
			if thumbErr := deletePairedThumbObject(bucket, key, mediaKind); thumbErr != nil {
				return deleted, skipped, thumbErr
			}
			if delBlobErr := deleteMediaBlobRow(ctx, uint64(blob.Id)); delBlobErr != nil {
				return deleted, skipped, delBlobErr
			}
		} else if decErr := decrementMediaBlobRefCount(ctx, uint64(blob.Id)); decErr != nil {
			return deleted, skipped, decErr
		}
	}
	return deleted, skipped, nil
}

func findMediaBlobByObjectKey(ctx context.Context, objectKey string) (*entity.UcgMediaBlob, error) {
	cols := dao.UcgMediaBlob.Columns()
	var blob entity.UcgMediaBlob
	err := dao.UcgMediaBlob.Ctx(ctx).Where(cols.ObjectKey, objectKey).Scan(&blob)
	if err != nil {
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "查询 blob 失败")
	}
	if blob.Id == 0 {
		return nil, nil
	}
	return &blob, nil
}

func decrementMediaBlobRefCount(ctx context.Context, blobID uint64) error {
	cols := dao.UcgMediaBlob.Columns()
	_, err := dao.UcgMediaBlob.Ctx(ctx).
		Where(cols.Id, blobID).
		WhereGT(cols.RefCount, 0).
		Decrement(cols.RefCount, 1)
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "递减 ref_count 失败")
	}
	return nil
}

func deleteMediaBlobRow(ctx context.Context, blobID uint64) error {
	cols := dao.UcgMediaBlob.Columns()
	_, err := dao.UcgMediaBlob.Ctx(ctx).Where(cols.Id, blobID).Delete()
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "删除 blob 行失败")
	}
	return nil
}
