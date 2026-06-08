package ucg

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"

	"hello/internal/dao"
	"hello/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

var sha256HexPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// MediaResolveResult resolve 查询结果。
type MediaResolveResult struct {
	Hit       bool
	ObjectKey string
	CdnURL    string
}

// ResolveMediaByHash 按 (contentHash, transformVersion) 查 blob 索引。
func ResolveMediaByHash(ctx context.Context, contentHash, transformVersion string, mediaKind int) (*MediaResolveResult, error) {
	hash := strings.ToLower(strings.TrimSpace(contentHash))
	version := strings.TrimSpace(transformVersion)
	if !sha256HexPattern.MatchString(hash) {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "contentHash 须为 64 位小写 hex SHA-256")
	}
	if version == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "transformVersion 无效")
	}
	if mediaKind != 1 && mediaKind != 2 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "mediaKind 须为 1 或 2")
	}

	cols := dao.UcgMediaBlob.Columns()
	var blob entity.UcgMediaBlob
	err := dao.UcgMediaBlob.Ctx(ctx).
		Where(cols.ContentHash, hash).
		Where(cols.TransformVersion, version).
		Scan(&blob)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &MediaResolveResult{Hit: false}, nil
		}
		return nil, gerror.WrapCode(gcode.CodeInternalError, err, "查询 blob 索引失败")
	}
	if blob.Id == 0 || strings.TrimSpace(blob.ObjectKey) == "" {
		return &MediaResolveResult{Hit: false}, nil
	}
	return &MediaResolveResult{
		Hit:       true,
		ObjectKey: blob.ObjectKey,
		CdnURL:    BuildCdnURL(blob.ObjectKey),
	}, nil
}
