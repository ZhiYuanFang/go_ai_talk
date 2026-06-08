package ucg

import (
	"context"

	"hello/internal/dao"
)

func mediaUploadOwnedBy(ctx context.Context, wxID int64, objectKey string) (bool, error) {
	cols := dao.UcgMediaUpload.Columns()
	n, err := dao.UcgMediaUpload.Ctx(ctx).
		Where(cols.WxId, wxID).
		Where(cols.ObjectKey, objectKey).
		Count()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func deleteMediaUploadRow(ctx context.Context, wxID int64, objectKey string) error {
	cols := dao.UcgMediaUpload.Columns()
	_, err := dao.UcgMediaUpload.Ctx(ctx).
		Where(cols.WxId, wxID).
		Where(cols.ObjectKey, objectKey).
		Delete()
	return err
}

func objectKeyReferencedByPost(ctx context.Context, objectKey string) (bool, error) {
	cols := dao.UcgPostMedia.Columns()
	n, err := dao.UcgPostMedia.Ctx(ctx).Where(cols.ObjectKey, objectKey).Count()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
