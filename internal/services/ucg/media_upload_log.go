package ucg

import (
	"context"
	"strings"
	"time"

	"hello/internal/dao"

	"github.com/gogf/gf/v2/frame/g"
)

// LogMediaUpload records ownership after presign or server-side upload succeeds.
func LogMediaUpload(ctx context.Context, wxID int64, objectKey string, mediaKind int) error {
	key := strings.TrimSpace(objectKey)
	if wxID <= 0 || key == "" {
		return nil
	}
	if mediaKind != 1 && mediaKind != 2 {
		mediaKind = 1
	}
	cols := dao.UcgMediaUpload.Columns()
	n, err := dao.UcgMediaUpload.Ctx(ctx).Where(cols.ObjectKey, key).Count()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	now := time.Now().Unix()
	_, err = dao.UcgMediaUpload.Ctx(ctx).Data(g.Map{
		cols.WxId:      wxID,
		cols.ObjectKey: key,
		cols.MediaKind: mediaKind,
		cols.CreatedAt: now,
	}).Insert()
	return err
}
