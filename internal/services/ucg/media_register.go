package ucg

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// RegisterMediaRequest register 请求体。
type RegisterMediaRequest struct {
	ObjectKey        string
	ContentHash      string
	TransformVersion string
	MediaKind        int
	DedupHit         bool
}

// RegisterMediaResult register 响应。
type RegisterMediaResult struct {
	ObjectKey string
	CdnURL    string
}

// RegisterMedia 登记 blob 与 ownership（miss 或 dedup hit）。
func RegisterMedia(ctx context.Context, wxID int64, req RegisterMediaRequest) (*RegisterMediaResult, error) {
	if wxID <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "wxId 无效")
	}
	hash := strings.ToLower(strings.TrimSpace(req.ContentHash))
	version := strings.TrimSpace(req.TransformVersion)
	objectKey := strings.TrimSpace(req.ObjectKey)
	if !sha256HexPattern.MatchString(hash) {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "contentHash 须为 64 位小写 hex SHA-256")
	}
	if version == "" || objectKey == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "参数无效")
	}
	if req.MediaKind != 1 && req.MediaKind != 2 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "mediaKind 须为 1 或 2")
	}

	var resultObjectKey string
	err := dao.UcgMediaBlob.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		cols := dao.UcgMediaBlob.Columns()
		var blob entity.UcgMediaBlob
		scanErr := tx.Model(dao.UcgMediaBlob.Table()).
			Where(cols.ContentHash, hash).
			Where(cols.TransformVersion, version).
			Scan(&blob)
		if scanErr != nil && !errors.Is(scanErr, sql.ErrNoRows) {
			return gerror.WrapCode(gcode.CodeInternalError, scanErr, "查询 blob 失败")
		}

		if req.DedupHit {
			if blob.Id == 0 {
				return gerror.NewCode(gcode.CodeInvalidParameter, "blob 不存在")
			}
			if blob.ObjectKey != objectKey {
				return gerror.NewCode(gcode.CodeInvalidParameter, "objectKey 与 blob 不匹配")
			}
			if _, err := tx.Model(dao.UcgMediaBlob.Table()).
				Where(cols.Id, blob.Id).
				Increment(cols.RefCount, 1); err != nil {
				return gerror.WrapCode(gcode.CodeInternalError, err, "递增 ref_count 失败")
			}
			resultObjectKey = blob.ObjectKey
			return upsertMediaUploadOwnershipTx(ctx, tx, wxID, blob.ObjectKey, req.MediaKind)
		}

		// miss path: blob must not exist yet (or treat concurrent insert as hit).
		if blob.Id > 0 {
			if blob.ObjectKey != objectKey {
				objectKey = blob.ObjectKey
			}
			if _, err := tx.Model(dao.UcgMediaBlob.Table()).
				Where(cols.Id, blob.Id).
				Increment(cols.RefCount, 1); err != nil {
				return gerror.WrapCode(gcode.CodeInternalError, err, "递增 ref_count 失败")
			}
			resultObjectKey = blob.ObjectKey
			return upsertMediaUploadOwnershipTx(ctx, tx, wxID, blob.ObjectKey, req.MediaKind)
		}

		exists, existErr := ossObjectExists(ctx, objectKey)
		if existErr != nil {
			return existErr
		}
		if !exists {
			return gerror.NewCode(gcode.CodeInvalidParameter, "OSS 对象不存在")
		}

		now := time.Now()
		_, insertErr := tx.Model(dao.UcgMediaBlob.Table()).Data(g.Map{
			cols.ContentHash:      hash,
			cols.TransformVersion: version,
			cols.ObjectKey:        objectKey,
			cols.MediaKind:        req.MediaKind,
			cols.RefCount:         1,
			cols.CreatedAt:        now,
		}).Insert()
		if insertErr != nil {
			if isDuplicateKeyError(insertErr) {
				if reloadErr := tx.Model(dao.UcgMediaBlob.Table()).
					Where(cols.ContentHash, hash).
					Where(cols.TransformVersion, version).
					Scan(&blob); reloadErr != nil || blob.Id == 0 {
					return gerror.WrapCode(gcode.CodeInternalError, insertErr, "并发 register 冲突")
				}
				resultObjectKey = blob.ObjectKey
				if _, err := tx.Model(dao.UcgMediaBlob.Table()).
					Where(cols.Id, blob.Id).
					Increment(cols.RefCount, 1); err != nil {
					return gerror.WrapCode(gcode.CodeInternalError, err, "递增 ref_count 失败")
				}
				return upsertMediaUploadOwnershipTx(ctx, tx, wxID, blob.ObjectKey, req.MediaKind)
			}
			return gerror.WrapCode(gcode.CodeInternalError, insertErr, "insert blob 失败")
		}
		resultObjectKey = objectKey
		return upsertMediaUploadOwnershipTx(ctx, tx, wxID, objectKey, req.MediaKind)
	})
	if err != nil {
		return nil, err
	}
	if req.MediaKind == 1 {
		if thumbErr := EnsureImageThumb(ctx, resultObjectKey); thumbErr != nil {
			return nil, thumbErr
		}
	}
	return &RegisterMediaResult{
		ObjectKey: resultObjectKey,
		CdnURL:    BuildCdnURL(resultObjectKey),
	}, nil
}

func upsertMediaUploadOwnershipTx(ctx context.Context, tx gdb.TX, wxID int64, objectKey string, mediaKind int) error {
	cols := dao.UcgMediaUpload.Columns()
	n, err := tx.Model(dao.UcgMediaUpload.Table()).
		Where(cols.WxId, wxID).
		Where(cols.ObjectKey, objectKey).
		Count()
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "查询 ownership 失败")
	}
	if n > 0 {
		return nil
	}
	now := time.Now().Unix()
	_, err = tx.Model(dao.UcgMediaUpload.Table()).Data(g.Map{
		cols.WxId:      wxID,
		cols.ObjectKey: objectKey,
		cols.MediaKind: mediaKind,
		cols.CreatedAt: now,
	}).Insert()
	if err != nil && isDuplicateKeyError(err) {
		return nil
	}
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err, "insert ownership 失败")
	}
	return nil
}

func ossObjectExists(ctx context.Context, objectKey string) (bool, error) {
	cfg := LoadOSSConfig(ctx)
	if err := validateOSSConfig(cfg); err != nil {
		return false, err
	}
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return false, gerror.WrapCode(gcode.CodeInternalError, err, "OSS 客户端初始化失败")
	}
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return false, gerror.WrapCode(gcode.CodeInternalError, err, "OSS Bucket 不可用")
	}
	ok, err := bucket.IsObjectExist(objectKey)
	if err != nil {
		return false, gerror.WrapCode(gcode.CodeInternalError, err, "OSS HEAD 失败")
	}
	return ok, nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	if err == sql.ErrNoRows {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "1062")
}
