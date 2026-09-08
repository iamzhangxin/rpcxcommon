package errors

import stderrors "errors"

// FromStorageWrite 将数据库唯一约束冲突转换为资源冲突并保留底层错误链。
func FromStorageWrite(err error) error {
	if err == nil {
		return nil
	}
	var state interface {
		// SQLState 遵循数据库驱动已有接口返回标准状态码。
		SQLState() string
	}
	if stderrors.As(err, &state) && state.SQLState() == "23505" {
		return ErrConflict.WithCause(err)
	}
	return ErrWriteFailed.WithCause(err)
}
