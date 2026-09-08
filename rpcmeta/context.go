package rpcmeta

import (
	"context"
	"strings"

	"github.com/bytedance/gopkg/cloud/metainfo"
)

const UserIdKey = "x-user-id"

// UserId 从持续透传的 Rpc 元信息读取已由入口校验的用户身份。
func UserId(ctx context.Context) string {
	id, _ := metainfo.GetPersistentValue(ctx, UserIdKey)
	return strings.TrimSpace(id)
}

// WithUserId 写入持续透传的用户身份，空值清除当前上下文的身份。
func WithUserId(ctx context.Context, id string) context.Context {
	id = strings.TrimSpace(id)
	if id == "" {
		return metainfo.DelPersistentValue(ctx, UserIdKey)
	}
	return metainfo.WithPersistentValue(ctx, UserIdKey, id)
}
