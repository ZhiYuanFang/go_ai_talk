package simuser

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// rollbackSimRegistration T1 失败时补偿：注销 sim wx 并删除 credential（幂等）。
func rollbackSimRegistration(ctx context.Context, wxID int64) {
	if wxID <= 0 {
		return
	}
	path := fmt.Sprintf("/device/internal/api/sim/wx/%d/deactivate", wxID)
	if err := deviceInternalPost(ctx, path, g.Map{}, nil); err != nil {
		glog.Warningf(ctx, "[simuser] rollback failed wxId=%d err=%v", wxID, err)
	}
	if err := DeleteWxCredentialByWxID(ctx, wxID); err != nil {
		glog.Warningf(ctx, "[simuser] rollback credential delete failed wxId=%d err=%v", wxID, err)
	}
}
