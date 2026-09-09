package errors

import (
	"context"
	stderrors "errors"

	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/klog"
)

// ToRpc 统一转换业务错误并保留取消语义和六位下游业务状态。
func ToRpc(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if stderrors.Is(err, context.Canceled) || stderrors.Is(err, context.DeadlineExceeded) {
		return err
	}
	var defined *Error
	if stderrors.As(err, &defined) {
		if defined.Code() < 500000 {
			return kerrors.NewBizStatusError(defined.Code(), defined.Message())
		}
		klog.CtxErrorf(ctx, "请求失败: %v", defined)
	} else if biz, ok := kerrors.FromBizStatusError(err); ok && biz.BizStatusCode() >= 100000 && biz.BizStatusCode() <= 999999 {
		return biz
	} else {
		klog.CtxErrorf(ctx, "请求失败，未分类的错误类型 %T", err)
	}
	return kerrors.NewBizStatusError(ErrInternal.Code(), ErrInternal.Message())
}
